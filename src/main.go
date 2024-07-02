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
type Range struct {
	Min    string `json:"min"`
	Max    string `json:"max"`
	Status int    `json:"status"`
}

// Ranges struct to hold an array of ranges
type Ranges struct {
	Ranges []Range `json:"ranges"`
}

var (
	// Global Vars
	Version = "v0.5"

	CharNewLine   = "\n"
	CharReadline  = '\n'
	RangeKey      *big.Int
	PrivKeyMinInt = new(big.Int)
	PrivKeyMaxInt = new(big.Int)

	// Usado no modo 3
	StepValue = 1000
	// Wallets
	Wallets = make(map[string][]byte)
	// Create Context
	ctx, ctxCancel = context.WithCancel(context.Background())
	// Create a channel to receive results from workers
	resultChan = make(chan *big.Int)
	// Variavel to update last processed wallet address
	lastkey string
	numCPU  int
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
	rangeNumber := PromptRangeNumber(len(ranges.Ranges))
	// Pergunta sobre modos de usar
	modoSelecionado := PromptModos(4) // quantidade de modos
	carteirasalva := fmt.Sprintf("%d", rangeNumber)
	// função HandleModoSelecionado - onde busca o modo selecionado do usuario. // talvez criar a funcao de favoritar essa opções e iniciar automaticamente?
	privKeyInt := HandleModoSelecionado(modoSelecionado, ranges, rangeNumber, carteirasalva)

	// Load wallet addresses from JSON file
	err = LoadWallets(filepath.Join(rootDir, "data", "wallets.json"))
	if err != nil {
		log.Fatalf("Failed to load wallets: %v", err)
	}

	keysChecked := int64(0)
	startTime := time.Now()

	// Number of CPU cores to use
	fmt.Printf("CPUs detectados: %s"+CharNewLine, green(numCPU))

	// Ask the user for the number of cpus
	cpusNumber := PromptCPUNumber()

	// Create a channel to send private keys to workers
	privKeyChan := make(chan *big.Int, cpusNumber*100000)
	runtime.GOMAXPROCS(cpusNumber)

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

	go func() {
		// Ticker for periodic updates every 5 seconds
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case privKey := <-privKeyChan:
				lastkey = fmt.Sprintf("%064x", privKey)

			case <-ticker.C:
				go func() {
					elapsedTime := time.Since(startTime).Seconds()
					keysPerSecond := float64(keysChecked) / elapsedTime
					fmt.Printf("Chaves checadas: %s Chaves por segundo: %s"+CharNewLine, humanize.Comma(int64(keysChecked)), humanize.Comma(int64(keysPerSecond)))
					if modoSelecionado == 2 || modoSelecionado == 3 {
						saveUltimaKeyWallet("ultimaChavePorCarteira.txt", carteirasalva, lastkey)
					}
				}()

			case <-ctx.Done():
				return
			}
		}
	}()

	// Send private keys to the workers
	go func() {
		defer close(privKeyChan)
		for {
			privKeyCopy := new(big.Int).Set(privKeyInt)
			select {
			case privKeyChan <- privKeyCopy:
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
				keysChecked++
			case <-ctx.Done():
				return
			}
		}
	}()

	// Control CTRL+C
	SysCtrlSignal := make(chan os.Signal, 1)
	signal.Notify(SysCtrlSignal, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Wait for a result from any worker
	var foundAddress *big.Int
	var foundAddressString string
	select {
	case <-SysCtrlSignal:
		if modoSelecionado >= 2 {
			saveUltimaKeyWallet("ultimaChavePorCarteira.txt", carteirasalva, lastkey)
		}
		fmt.Println("Stop Program" + CharNewLine)
		ctxCancel()

	case foundAddress = <-resultChan:
		wif := btc_utils.GenerateWif(foundAddress)
		color.Yellow(CharNewLine+"Chave privada encontrada: %064x"+CharNewLine, foundAddress)
		color.Yellow("WIF: %s"+CharNewLine, wif)

		// Parar todas as Go Rotines
		ctxCancel()

		// Obter a data e horário atuais
		currentTime := time.Now().Format("2006-01-02 15:04:05")

		// Abrir ou criar o arquivo chaves_encontradas.txt
		file, err := os.OpenFile("chaves_encontradas.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Erro ao abrir o arquivo:", err)
		} else {
			_, err = file.WriteString(fmt.Sprintf("Data/Hora: %s | Chave privada: %064x | WIF: %s\n", currentTime, foundAddress, wif))
			if err != nil {
				fmt.Println("Erro ao escrever no arquivo:", err)
			} else {
				fmt.Println("Chaves salvas com sucesso.")
			}
			file.Close()
		}
		if modoSelecionado == 2 || modoSelecionado == 3 {
			foundAddressString = fmt.Sprintf("%064x", foundAddress)
			saveUltimaKeyWallet("ultimaChavePorCarteira.txt", carteirasalva, foundAddressString)
		}
	}

	// Wait for all workers to finish
	wg.Wait()

	elapsedTime := time.Since(startTime).Seconds()
	keysPerSecond := float64(keysChecked) / elapsedTime
	fmt.Printf("Chaves checadas: %s"+CharNewLine, humanize.Comma(int64(keysChecked)))
	fmt.Printf("Tempo: %.2f seconds"+CharNewLine, elapsedTime)
	fmt.Printf("Chaves por segundo: %s"+CharNewLine, humanize.Comma(int64(keysPerSecond)))
}

func randRange(min, max int) int {
	return rand.IntN(max-min) + min
}

// start na workers
func worker(ctx context.Context, privKeyChan <-chan *big.Int) {
	for privKeyInt := range privKeyChan {
		if ctx.Err() != nil {
			return
		}
		address := btc_utils.CreatePublicHash160(privKeyInt)
		if _, ok := Wallets[string(address)]; ok {
			resultChan <- privKeyInt
			return
		}
	}
}

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
