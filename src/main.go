package main

import (
	"bytes"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"btcgo/src/crypto/btc_utils"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
)

// Wallets struct to hold the array of wallet addresses
type Wallets struct {
	Addresses       [][]byte `json:"wallets"`
}

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

func titulo() {
	fmt.Println("\x1b[38;2;250;128;114m" + "╔═══════════════════════════════════════╗")
	fmt.Println("║\x1b[0m\x1b[36m" + "   ____ _______ _____    _____  ____   " + "\x1b[0m\x1b[38;2;250;128;114m" + "║")
	fmt.Println("║\x1b[0m\x1b[36m" + "  |  _ \\__   __/ ____|  / ____|/ __ \\  " + "\x1b[0m\x1b[38;2;250;128;114m" + "║")
	fmt.Println("║\x1b[0m\x1b[36m" + "  | |_) | | | | |      | |  __| |  | | " + "\x1b[0m\x1b[38;2;250;128;114m" + "║")
	fmt.Println("║\x1b[0m\x1b[36m" + "  |  _ <  | | | |      | | |_ | |  | | " + "\x1b[0m\x1b[38;2;250;128;114m" + "║")
	fmt.Println("║\x1b[0m\x1b[36m" + "  | |_) | | | | |____  | |__| | |__| | " + "\x1b[0m\x1b[38;2;250;128;114m" + "║")
	fmt.Println("║\x1b[0m\x1b[36m" + "  |____/  |_|  \\_____|  \\_____|\\____/  " + "\x1b[0m\x1b[38;2;250;128;114m" + "║")
	fmt.Println("║\x1b[0m\x1b[36m" + "                                       " + "\x1b[0m\x1b[38;2;250;128;114m" + "║")
	fmt.Println("╚════\x1b[32m" + "Investidor Internacional - v0.4" + "\x1b[0m\x1b[38;2;250;128;114m════╝" + "\x1b[0m")
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
	wallets, err := LoadWallets(filepath.Join(rootDir, "data", "wallets.json"))
	if err != nil {
		log.Fatalf("Failed to load wallets: %v", err)
	}
	wallet := wallets.Addresses[rangeNumber-1]
	// pergunta se deseja verificar todas as carteiras ou apenas uma por vez
	modoUniqueOrAll := PromptUniqueOrAll()

	// Pergunta sobre modos de usar
	modoSelecionado := PromptModos(2) // quantidade de modos

	var carteirasalva string
	carteirasalva = fmt.Sprintf("%d", rangeNumber)
	privKeyInt := new(big.Int)

	// função HandleModoSelecionado - onde busca o modo selecionado do usuario. // talvez criar a funcao de favoritar essa opções e iniciar automaticamente?
	privKeyInt = HandleModoSelecionado(modoSelecionado, ranges, rangeNumber, privKeyInt, carteirasalva)

	// Load wallet addresses from JSON file

	keysChecked := 0
	startTime := time.Now()

	// Number of CPU cores to use
	numCPU := runtime.NumCPU()
	fmt.Printf("CPUs detectados: %s\n", green(numCPU))
	runtime.GOMAXPROCS(numCPU)

	// Ask the user for the number of cpus
	cpusNumber := PromptCPUNumber(numCPU)

	// Create a channel to send private keys to workers
	privKeyChan := make(chan *big.Int, cpusNumber)
	// Create a channel to receive results from workers
	resultChan := make(chan *big.Int)
	// Create a wait group to wait for all workers to finish
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < cpusNumber; i++ {
		wg.Add(1)
		if modoUniqueOrAll == 2 {
			go workerForUniqueFind(wallet, privKeyChan, resultChan, &wg)
		} else {
			go worker(wallets, privKeyChan, resultChan, &wg)
		}
	}

	// Ticker for periodic updates every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	done := make(chan struct{})

	// Variavel to update last processed wallet address
	var lastkey string
	// Goroutine to update last processed wallet address
	go func() {
		for {
			select {
			case privKey := <-privKeyChan:
				lastkey = fmt.Sprintf("%064x", privKey)
			case <-done:
				return
			}
		}
	}()

	// Goroutine to print speed updates
	go func() {
		for {
			select {
			case <-ticker.C:
				elapsedTime := time.Since(startTime).Seconds()
				keysPerSecond := float64(keysChecked) / elapsedTime
				fmt.Printf("Chaves checadas: %s Chaves por segundo: %s\n", humanize.Comma(int64(keysChecked)), humanize.Comma(int64(keysPerSecond)))
				if modoSelecionado == 2 {
					saveUltimaKeyWallet("ultimaChavePorCarteira.txt", carteirasalva, lastkey)
				}
			case <-done:
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
				privKeyInt.Add(privKeyInt, big.NewInt(1))
				keysChecked++
			case <-done:
				return
			}
		}
	}()

	// Wait for a result from any worker
	var foundAddress *big.Int
	var foundAddressString string
	select {
	case foundAddress = <-resultChan:
		wif := btc_utils.GenerateWif(foundAddress)
		color.Yellow("Chave privada encontrada: %064x\n", foundAddress)
		color.Yellow("WIF: %s", wif)

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

		if modoSelecionado == 2 {
			foundAddressString = fmt.Sprintf("%064x", foundAddress)
			saveUltimaKeyWallet("ultimaChavePorCarteira.txt", carteirasalva, foundAddressString)
		}

		close(privKeyChan)

	}

	// Wait for all workers to finish
	wg.Wait()

	elapsedTime := time.Since(startTime).Seconds()
	keysPerSecond := float64(keysChecked) / elapsedTime
	fmt.Printf("Chaves checadas: %s\n", humanize.Comma(int64(keysChecked)))
	fmt.Printf("Tempo: %.2f seconds\n", elapsedTime)
	fmt.Printf("Chaves por segundo: %s\n", humanize.Comma(int64(keysPerSecond)))

}

// start na workers
func worker(wallets *Wallets, privKeyChan <-chan *big.Int, resultChan chan<- *big.Int, wg *sync.WaitGroup) {
	defer wg.Done()
	for privKeyInt := range privKeyChan {
		address := btc_utils.CreatePublicHash160(privKeyInt)
		if Contains(wallets.Addresses, address) {
			select {
			case resultChan <- privKeyInt:
				return
			default:
				return
			}
		}
	}
}
func workerForUniqueFind(wallet []byte, privKeyChan <-chan *big.Int, resultChan chan<- *big.Int, wg *sync.WaitGroup) {
	defer wg.Done()
	for privKeyInt := range privKeyChan {
		address := btc_utils.CreatePublicHash160(privKeyInt)
		if bytes.Equal(wallet, address) {
			select {
			case resultChan <- privKeyInt:
				return
			default:
				return
			}
		}
	}
}
