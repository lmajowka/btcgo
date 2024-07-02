package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"math/rand/v2"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"btcgo/src/crypto/btc_utils"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
)

// Range struct to hold the minimum, maximum, and status
type RangeStruct struct {
	Min    string `json:"min"`
	Max    string `json:"max"`
	Status int    `json:"status"`
}

// Ranges struct to hold an array of ranges
type Ranges struct {
	Range []RangeStruct `json:"ranges"`
}

type channelResultStruct struct {
	PrivKey big.Int
	Wallet  string
}

var (
	// Global Vars
	Version = "v0.5"

	CharNewLine   = "\n"
	CharReadline  = '\n'
	RangeKey      *big.Int
	PrivKeyMinInt = new(big.Int)
	PrivKeyMaxInt = new(big.Int)
	WallettoFind  = ""
	KeysChecked   = int64(0)

	Ticker *time.Ticker

	WorkerLastKey chan big.Int

	// Usado no modo 3
	StepValue = 100000
	// Wallets
	Wallets = make(map[string]string)
	// Create Context
	ctx, ctxCancel = context.WithCancel(context.Background())
	// Create a channel to receive results from workers
	resultChan = make(chan channelResultStruct, 1)
	// Variavel to update last processed wallet address
	lastkey   string
	numCPU    int
	startTime time.Time
)

// O GO chama esta func automáticamente sempre que inicializa
func init() {
	if runtime.GOOS == "windows" {
		CharNewLine = "\n\r"
		CharReadline = '\r'
	}
	numCPU = runtime.NumCPU()
}

func main() {
	green := color.New(color.FgGreen).SprintFunc()
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("Erro ao obter o caminho do executável: %v\n", err)
		return
	}
	rootDir := filepath.Dir(exePath)

	ranges, err := LoadRanges(filepath.Join(rootDir, "data", "ranges.json"))
	if err != nil {
		log.Fatalf("Failed to load ranges: %v", err)
	}

	ClearConsole()
	titulo()

	// Ask the user for the range number
	rangeNumber := PromptRangeNumber(len(ranges.Range))
	// Pergunta sobre modos de usar
	modoSelecionado := PromptModos(4) // quantidade de modos
	carteirasalva := fmt.Sprintf("%d", rangeNumber)
	// função HandleModoSelecionado - onde busca o modo selecionado do usuario. // talvez criar a funcao de favoritar essa opções e iniciar automaticamente?
	privKeyInt := HandleModoSelecionado(modoSelecionado, ranges, rangeNumber, carteirasalva)

	// Load wallet addresses from JSON file
	err = LoadWallets(filepath.Join(rootDir, "data", "wallets.json"), rangeNumber)
	if err != nil {
		log.Fatalf("Failed to load wallets: %v", err)
	}

	// Number of CPU cores to use
	fmt.Printf("CPUs detectados: %s"+CharNewLine, green(numCPU))

	// Ask the user for the number of cpus
	cpusNumber := PromptCPUNumber()

	// Create a channel to send keys to workers
	privKeyChan := make(chan big.Int, cpusNumber)
	defer close(privKeyChan)
	// Channel utilizado para actualizar a ultima chave processada
	WorkerLastKey = make(chan big.Int, 1)
	defer close(WorkerLastKey)

	// Set Max CPU
	runtime.GOMAXPROCS(cpusNumber)

	// Inicializa Rotina de gravação
	go saveFindKey(ctx, modoSelecionado, carteirasalva)

	// Create a wait group to wait for all workers to finish
	var wg sync.WaitGroup
	// Start worker goroutines
	for i := 0; i < cpusNumber; i++ {
		wg.Add(1)
		go func(cx context.Context) {
			defer wg.Done()
			worker(cx, privKeyChan)
		}(ctx)
	}

	// Control CTRL+C
	SysCtrlSignal := make(chan os.Signal, 1)
	signal.Notify(SysCtrlSignal, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Set Start Time
	startTime = time.Now()

	// Ticker for periodic updates every 5 seconds
	Ticker = time.NewTicker(5 * time.Second)
	defer Ticker.Stop()

	// Mostra qual a wallet que procura
	fmt.Println("A procurar chave para a carteira", WallettoFind)

	// Inicializa a rotina que cria as chaves
	go createKeys(ctx, modoSelecionado, privKeyChan, privKeyInt)

	// Controla os channels de informações
	//go func() {
	for {
		select {
		case <-ctx.Done():
			return

		case privKey := <-WorkerLastKey:
			lastkey = fmt.Sprintf("%064x", &privKey)

		case <-Ticker.C:
			elapsedTime := time.Since(startTime).Seconds()
			keysPerSecond := float64(KeysChecked) / elapsedTime
			fmt.Printf("Chaves checadas: %s Chaves por segundo: %s"+CharNewLine, humanize.Comma(int64(KeysChecked)), humanize.Comma(int64(keysPerSecond)))
			//fmt.Println("Ultima Chave", lastkey)
			if modoSelecionado <= 3 {
				saveUltimaKeyWallet("ultimaChavePorCarteira.json", carteirasalva, lastkey)
			}

		case <-SysCtrlSignal:
			Ticker.Stop()
			if modoSelecionado <= 3 {
				saveUltimaKeyWallet("ultimaChavePorCarteira.json", carteirasalva, lastkey)
			}
			fmt.Println("Stop Program" + CharNewLine)
			ctxCancel()
		}

		if ctx.Err() != nil {
			// Force Run Worker to see ctx.err
			for i := 0; i < cpusNumber; i++ {
				privKeyChan <- *big.NewInt(0)
			}
			break
		}
	}
	//}()

	// Wait for all workers to finish
	wg.Wait()

	elapsedTime := time.Since(startTime).Seconds()
	keysPerSecond := float64(KeysChecked) / elapsedTime
	fmt.Printf("Chaves checadas: %s"+CharNewLine, humanize.Comma(int64(KeysChecked)))
	fmt.Printf("Tempo: %.2f seconds"+CharNewLine, elapsedTime)
	fmt.Printf("Chaves por segundo: %s"+CharNewLine, humanize.Comma(int64(keysPerSecond)))
}

// Cria e envia as chaves para os workers
func createKeys(ctx context.Context, modoSelecionado int, privKeyChan chan big.Int, privKeyInt *big.Int) {
	for {
		privKeyCopy := new(big.Int).Set(privKeyInt)
		select {
		case privKeyChan <- *privKeyCopy:
			if modoSelecionado == 3 {
				z := randRange(1, StepValue)
				y := randRange(1, 10)
				if y >= 5 {
					privKeyInt.Add(privKeyInt, big.NewInt(int64(z)))
				} else {
					privKeyInt.Sub(privKeyInt, big.NewInt(int64(z)))
				}
			} else if modoSelecionado == 4 {
				key := CalcPrivKey(big.NewFloat(randFloat(1, 99) / 100))
				privKeyInt.SetString(key, 16)
			} else {
				privKeyInt.Add(privKeyInt, big.NewInt(1))
			}
			WorkerLastKey <- *privKeyInt
			KeysChecked++
		case <-ctx.Done():
			return
		}
	}
}

// Grava as chaves que encontra
func saveFindKey(ctx context.Context, modoSelecionado int, carteirasalva string) {
	for {
		select {
		case <-ctx.Done():
			return

		case result := <-resultChan:
			// Para o Ticker para evitar que grave dados ao mesmo tempo
			Ticker.Stop()
			// Cria Dados sobre a chave
			wif := btc_utils.GenerateWif(&result.PrivKey)
			color.Yellow("Wallet: %s", result.Wallet)
			color.Yellow("Chave privada encontrada: %064x", &result.PrivKey)
			color.Yellow("WIF: %s", wif)

			// Obter a data e horário atuais
			currentTime := time.Now().Format("2006-01-02 15:04:05")

			// Abrir ou criar o arquivo chaves_encontradas.txt
			file, err := os.OpenFile("chaves_encontradas.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Println("Erro ao abrir o arquivo:", err)
			} else {
				_, err = file.WriteString(fmt.Sprintf("Data/Hora: %s | Chave privada: %064x | WIF: %s | Wallet %s\n", currentTime, &result.PrivKey, wif, result.Wallet))
				if err != nil {
					fmt.Println("Erro ao escrever no arquivo:", err)
				} else {
					fmt.Println("Chaves salvas com sucesso.")
				}
				file.Close()
			}
			// Grava dados
			if modoSelecionado <= 3 {
				foundAddressString := fmt.Sprintf("%064x", &result.PrivKey)
				saveUltimaKeyWallet("ultimaChavePorCarteira.json", carteirasalva, foundAddressString)
			}
			// Verifica se terminou
			if WallettoFind == result.Wallet {
				ctxCancel()
				return
			} else {
				// Start Ticker Again
				Ticker = time.NewTicker(5 * time.Second)
			}
		}
	}
}

// start workers
func worker(ctx context.Context, privKeyChan chan big.Int) {
	for privKeyInt := range privKeyChan {
		if ctx.Err() != nil {
			return
		}
		address := btc_utils.Hash160ToAddress(btc_utils.CreatePublicHash160(&privKeyInt))
		if _, ok := Wallets[address]; ok {
			if WallettoFind != address {
				fmt.Println("Chave encontrada para a Wallet :", address)
				fmt.Println("Não é a Wallet que procura " + WallettoFind + ", Continua Procura...")
				resultChan <- channelResultStruct{
					PrivKey: privKeyInt,
					Wallet:  address,
				}
			} else {
				fmt.Println("Chave encontrada para a Wallet :", Wallets[address])
				resultChan <- channelResultStruct{
					PrivKey: privKeyInt,
					Wallet:  address,
				}
				return
			}
		}
	}
}

// Calcula um random para o step
func randRange(min, max int) int {
	return rand.IntN(max-min) + min
}

// Titulo Client
func titulo() {
	fmt.Println("\x1b[38;2;250;128;114m" + "╔═══════════════════════════════════════╗")
	fmt.Println("║\x1b[0m\x1b[36m" + "   ____ _______ _____    _____  ____   " + "\x1b[0m\x1b[38;2;250;128;114m" + "║")
	fmt.Println("║\x1b[0m\x1b[36m" + "  |  _ \\__   __/ ____|  / ____|/ __ \\  " + "\x1b[0m\x1b[38;2;250;128;114m" + "║")
	fmt.Println("║\x1b[0m\x1b[36m" + "  | |_) | | | | |      | |  __| |  | | " + "\x1b[0m\x1b[38;2;250;128;114m" + "║")
	fmt.Println("║\x1b[0m\x1b[36m" + "  |  _ <  | | | |      | | |_ | |  | | " + "\x1b[0m\x1b[38;2;250;128;114m" + "║")
	fmt.Println("║\x1b[0m\x1b[36m" + "  | |_) | | | | |____  | |__| | |__| | " + "\x1b[0m\x1b[38;2;250;128;114m" + "║")
	fmt.Println("║\x1b[0m\x1b[36m" + "  |____/  |_|  \\_____|  \\_____|\\____/  " + "\x1b[0m\x1b[38;2;250;128;114m" + "║")
	fmt.Println("║\x1b[0m\x1b[36m" + "                                       " + "\x1b[0m\x1b[38;2;250;128;114m" + "║")
	fmt.Println("╚════\x1b[32m" + "Investidor Internacional - " + Version + "\x1b[0m\x1b[38;2;250;128;114m════╝" + "\x1b[0m")
}

// Clear Console
func ClearConsole() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}
