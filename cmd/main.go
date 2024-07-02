package main

import (
	"btcgo/internal/core"
	"btcgo/internal/utils"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
)

func main() {
	green := color.New(color.FgGreen).SprintFunc()

	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("Erro ao obter o caminho do executável: %v\n", err)
		return
	}
	rootDir := filepath.Dir(exePath)

	ranges, err := utils.LoadRanges(filepath.Join(rootDir, "data", "ranges.json"))
	if err != nil {
		log.Fatalf("Failed to load ranges: %v", err)
	}

	color.Cyan("BTC GO - Investidor Internacional")
	color.Yellow("v0.4")

	// Ask the user for the range number
	rangeNumber := utils.PromptRangeNumber(len(ranges.Ranges))

	// Pergunta sobre modos de usar
	modoSelecionado := utils.PromptMods(2) // quantidade de modos

	var carteirasalva string
	carteirasalva = fmt.Sprintf("%d", rangeNumber)
	privKeyInt := new(big.Int)

	// função HandleModoSelecionado - onde busca o modo selecionado do usuario. // talvez criar a funcao de favoritar essa opções e iniciar automaticamente?
	privKeyInt = utils.HandleModSelected(modoSelecionado, ranges, rangeNumber, privKeyInt, carteirasalva)

	// Load wallet addresses from JSON file
	wallets, err := utils.LoadWallets(filepath.Join(rootDir, "data", "wallets.json"))
	if err != nil {
		log.Fatalf("Failed to load wallets: %v", err)
	}

	var keysChecked int64 = 0
	startTime := time.Now()

	// Number of CPU cores to use
	numCPU := runtime.NumCPU()
	fmt.Printf("CPUs detectados: %s\n", green(numCPU))
	runtime.GOMAXPROCS(numCPU)

	// Ask the user for the number of cpus
	cpusNumber := utils.ReadCPUsForUse()

	// Create a channel to send private keys to workers
	privKeyChan := make(chan *big.Int, cpusNumber)
	// Create a channel to receive results from workers
	resultChan := make(chan *big.Int)
	// Create a wait group to wait for all workers to finish
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < cpusNumber; i++ {
		wg.Add(1)
		go core.Worker(wallets, privKeyChan, resultChan, &wg)
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
					_ = utils.SaveLastKeyWallet("ultimaChavePorCarteira.txt", carteirasalva, lastkey)
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
		wif := utils.GenerateWif(foundAddress)
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
			_ = utils.SaveLastKeyWallet("ultimaChavePorCarteira.txt", carteirasalva, foundAddressString)
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
