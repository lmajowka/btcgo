package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"os"
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
	Addresses [][]byte `json:"wallets"`
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

	color.Cyan("BTC GO - Investidor Internacional")
	color.Yellow("v0.4")

	// Ask the user for the range number
	rangeNumber := PromptRangeNumber(len(ranges.Ranges))

	// Pergunta sobre modos de usar
	modoSelecionado := PromptModos(3) // quantidade de modos

	var carteirasalva string
	carteirasalva = fmt.Sprintf("%d", rangeNumber)
	privKeyInt := new(big.Int)

	// função HandleModoSelecionado - onde busca o modo selecionado do usuario. // talvez criar a funcao de favoritar essa opções e iniciar automaticamente?
	if modoSelecionado < 3 {
		privKeyInt = HandleModoSelecionado(modoSelecionado, ranges, rangeNumber, privKeyInt, carteirasalva)
	}

	privKeyInt = HandleModoSelecionado(modoSelecionado, ranges, rangeNumber, privKeyInt, carteirasalva)

	// Load wallet addresses from JSON file
	wallets, err := LoadWallets(filepath.Join(rootDir, "data", "wallets.json"))
	if err != nil {
		log.Fatalf("Failed to load wallets: %v", err)
	}

	// Marton - Vamos verificar se o range inicial é menor que o range final, exibir o range e o total de combinações:
	rangeMinHex := ranges.Ranges[rangeNumber-1].Min
	rangeMaxHex := ranges.Ranges[rangeNumber-1].Max
	rangeMinInt := new(big.Int)
	rangeMinInt.SetString(rangeMinHex[2:], 16)
	rangeMaxInt := new(big.Int)
	rangeMaxInt.SetString(rangeMaxHex[2:], 16)
	combinacoes := new(big.Int).Sub(rangeMaxInt, rangeMinInt)
	combinacoesFloat, _ := new(big.Float).SetInt(combinacoes).Float64() // Converte combinacoes para float64
	fmt.Println("Range.Min: " + rangeMinInt.Text(16))                   // Imprime em Hexadecimal
	fmt.Println("Range.Max: " + rangeMaxInt.Text(16))                   // Imprime em Hexadecimal
	fmt.Println("Total de Combinações: " + combinacoes.Text(10))        // Imprime em base decimal
	if rangeMinInt.Cmp(rangeMaxInt) > 0 {
		fmt.Println("Erro: o range inicial não pode ser maior que o range final.")
		os.Exit(1)
	}

	// Caso tenha escolhido o modo aleatório, vamos obter os valores iniciais:
	if modoSelecionado == 3 {
		updatePrivKeyInt(privKeyInt, rangeMinInt, rangeMaxInt, modoSelecionado)
	}

	fmt.Println("Posição Inicial: " + privKeyInt.Text(16)) // Imprime em Hexadecimal

	keysChecked := 0
	startTime := time.Now()

	// Number of CPU cores to use
	numCPU := runtime.NumCPU()
	fmt.Printf("CPUs detectados: %s\n", green(numCPU))
	runtime.GOMAXPROCS(numCPU)

	// Ask the user for the number of cpus
	cpusNumber := PromptCPUNumber()

	// Create a channel to send private keys to workers
	privKeyChan := make(chan *big.Int, cpusNumber)
	// Create a channel to receive results from workers
	resultChan := make(chan *big.Int)
	// Create a wait group to wait for all workers to finish
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < cpusNumber; i++ {
		wg.Add(1)
		go worker(wallets, privKeyChan, resultChan, &wg)
	}

	// Ticker for periodic updates every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	done := make(chan struct{})

	// Ticker para obter um número aleatório dentro do range a cada hora:
	tickerRandom := time.NewTicker(30 * time.Minute)
	defer tickerRandom.Stop()

	// Variavel to update last processed wallet address
	var lastkey string

	// Vamos obter uma nova posição aleatória a cada hora:
	go func() {
		for range tickerRandom.C {
			updatePrivKeyInt(privKeyInt, rangeMinInt, rangeMaxInt, modoSelecionado)
		}
	}()

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
				remainingKeys := combinacoesFloat - float64(keysChecked)                    // Calcula o número de chaves restantes a serem verificadas
				estimatedTime := remainingKeys / keysPerSecond                              // Calcula o tempo estimado para conclusão em segundos
				posicaoPercent := calculatePercentage(privKeyInt, rangeMinInt, rangeMaxInt) // Calcula a posição, em porcentagem
				posicaoPercentStr := fmt.Sprintf("%.12f", posicaoPercent)                   // Formata com 12 casas decimais
				fmt.Printf("%s - Posição: 0x%s (%s%%) ; Chaves checadas: %s ; Chaves por segundo: %s ; Tempo restante: %s\n", time.Now().Format("2006-01-02 15:04:05"), privKeyInt.Text(16), posicaoPercentStr, humanize.Comma(int64(keysChecked)), humanize.Comma(int64(keysPerSecond)), formatDuration(estimatedTime))
				if modoSelecionado == 2 {
					saveUltimaKeyWallet("ultimaChavePorCarteira.txt", carteirasalva, lastkey)
				}

				// Marton - Verifica se a chave atual é maior que o range final:
				if privKeyInt.Cmp(rangeMaxInt) > 0 {
					fmt.Println("O privKeyInt atual é maior que o range final. Vamos calcular nova posição aleatória.")
					updatePrivKeyInt(privKeyInt, rangeMinInt, rangeMaxInt, modoSelecionado)
				}

				// Marton - Verifica se a chave atual é menor que o range inicial:
				if privKeyInt.Cmp(rangeMinInt) < 0 {
					fmt.Println("O privKeyInt atual é menor que o range inicial. Vamos calcular nova posição aleatória.")
					updatePrivKeyInt(privKeyInt, rangeMinInt, rangeMaxInt, modoSelecionado)
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

// Essa função calcula a posição privKeyInt representa entre rangeMinInt e rangeMaxInt, em porcentagem
func calculatePercentage(pos, min, max *big.Int) float64 {
	// Calcula o intervalo total
	totalRange := new(big.Int).Sub(max, min)

	// Calcula a posição relativa de privKeyInt dentro do intervalo
	relativePosition := new(big.Int).Sub(pos, min)

	// Converte totalRange e relativePosition para float64
	totalRangeFloat, _ := new(big.Float).SetInt(totalRange).Float64()
	relativePositionFloat, _ := new(big.Float).SetInt(relativePosition).Float64()

	// Calcula a porcentagem
	percentage := (relativePositionFloat / totalRangeFloat) * 100

	return percentage - 1
}

// Marton - A função recebe um tempo estimado para conclusão e retorna string com anos;
func formatDuration(seconds float64) string {
	// Converte os segundos em um valor inteiro
	totalSeconds := int64(seconds)

	// Define constantes para as durações
	const (
		secondsInMinute = 60
		secondsInHour   = 60 * secondsInMinute
		secondsInDay    = 24 * secondsInHour
		secondsInMonth  = 30 * secondsInDay
		secondsInYear   = 12 * secondsInMonth
	)

	years := totalSeconds / secondsInYear
	totalSeconds %= secondsInYear

	/*
		months := totalSeconds / secondsInMonth
		totalSeconds %= secondsInMonth

		days := totalSeconds / secondsInDay
		totalSeconds %= secondsInDay
	*/

	// Monta a string formatada
	formattedDuration := fmt.Sprintf("%d anos" /*, %d mes(es), %d dia(s)"*/, years /*, months, days*/)
	return formattedDuration
}

func updatePrivKeyInt(privKeyInt, rangeMinInt, rangeMaxInt *big.Int, modoSelecionado int) {
	rangeSize := new(big.Int).Sub(rangeMaxInt, rangeMinInt)
	randomNum, err := rand.Int(rand.Reader, rangeSize)
	if err != nil {
		fmt.Println("Erro ao gerar número aleatório:", err)
		return
	}
	privKeyInt.Add(randomNum, rangeMinInt)

	posicaoPercent := calculatePercentage(privKeyInt, rangeMinInt, rangeMaxInt) // Calcula a posição, em porcentagem
	posicaoPercentStr := fmt.Sprintf("%.12f", posicaoPercent)                   // Formata com 12 casas decimais

	// Representação gráfica da porcentagem:
	graphicRepresentation := generateGraphicRepresentation(posicaoPercent)

	fmt.Printf("\nMovendo para nova posição aleatória: %s ; Nova posição: %s  -  %s\n", privKeyInt.Text(16), posicaoPercentStr, graphicRepresentation)

	if modoSelecionado == 3 {
		if posicaoPercent < 4 {
			fmt.Println("Não queremos processar os primeiros 4% do range. Vamos re-calcular nova posição aleatória.")
			updatePrivKeyInt(privKeyInt, rangeMinInt, rangeMaxInt, modoSelecionado)
		}
		if posicaoPercent > 99 {
			fmt.Println("Não queremos processar os últimos 1% do range. Vamos re-calcular nova posição aleatória.")
			updatePrivKeyInt(privKeyInt, rangeMinInt, rangeMaxInt, modoSelecionado)
		}
	}

	fmt.Println()

}

// Aqui eu gero uma representação gráfica, em ASCII, da porcentagem:
func generateGraphicRepresentation(percentage float64) string {
	totalLength := 50
	position := int(percentage * float64(totalLength) / 100)

	if position >= totalLength {
		position = totalLength - 1
	} else if position < 0 {
		position = 0
	}

	representation := make([]rune, totalLength)
	for i := range representation {
		representation[i] = '_'
	}
	representation[position] = 'X'

	return fmt.Sprintf("[%s]", string(representation))
}
