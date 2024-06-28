package main

import (
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
	"strconv"
	"strings"


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
	color.Yellow("v0.3")

	// Ask the user for the range number
	rangeNumber := PromptRangeNumber(len(ranges.Ranges))


	// Pergunta sobre modos de usar
	modoSelecionado := PromptModos(2) // quantidade de modos



	var carteirasalva string
	carteirasalva = fmt.Sprintf("%d", rangeNumber)

	privKeyInt := new(big.Int)


	if(modoSelecionado == 1){
		// Initialize privKeyInt with the minimum value of the selected range
		privKeyHex := ranges.Ranges[rangeNumber-1].Min
		privKeyInt.SetString(privKeyHex[2:], 16)	
	}else if(modoSelecionado == 2){
		// Carrega a última chave privada salva para a carteira específica
		verificaKey, err := LoadUltimaKeyWallet("ultimaChavePorCarteira.txt", carteirasalva)
		if err != nil || verificaKey == "" {
			// FAZER PERGUNTA SE DESEJA INFORMAR O NUMERO DE INCIO DO MODO SEQUENCIAL OU COMEÇAR DO INICIO

			msSequencialouInicio := PromptAuto("Opção 1: Deseja começar do inicio da busca (não efetivo) ou \nOpção 2: Escolher entre o range(porcentagem) da carteira informada? \nPor favor numero entre 1 ou 2:",2)
			if(msSequencialouInicio == 2){
				
			// Definindo as variáveis privKeyMinInt e privKeyMaxInt como big.Int
			privKeyMinInt := new(big.Int)
			privKeyMaxInt := new(big.Int)		
			privKeyMin := ranges.Ranges[rangeNumber-1].Min
			privKeyMax := ranges.Ranges[rangeNumber-1].Max
			privKeyMinInt.SetString(privKeyMin[2:], 16)
			privKeyMaxInt.SetString(privKeyMax[2:], 16)

			// Calculando a diferença entre privKeyMaxInt e privKeyMinInt
			rangeKey := new(big.Int).Sub(privKeyMaxInt, privKeyMinInt)

			// Solicitando a porcentagem do range da carteira como entrada
			var rangeCarteiraSequencialStr string
			fmt.Print("Informe a porcentagem do range da carteira entre 1 a 100: ")
			fmt.Scanln(&rangeCarteiraSequencialStr)

			// Substituindo vírgulas por pontos se necessário
			rangeCarteiraSequencialStr = strings.Replace(rangeCarteiraSequencialStr, ",", ".", -1)

			// Convertendo a porcentagem para um número decimal
			rangeCarteiraSequencial, err := strconv.ParseFloat(rangeCarteiraSequencialStr, 64)
			if err != nil {
				fmt.Println("Erro ao ler porcentagem:", err)
				return
			}

			// Verificando se a porcentagem está no intervalo válido
			if rangeCarteiraSequencial < 1 || rangeCarteiraSequencial > 100 {
				fmt.Println("Porcentagem fora do intervalo válido (1 a 100).")
				return
			}

			// Calculando o valor de rangeKey multiplicado pela porcentagem
			rangeMultiplier := new(big.Float).Mul(new(big.Float).SetInt(rangeKey), big.NewFloat(rangeCarteiraSequencial/100.0))

			// Convertendo o resultado para inteiro (arredondamento para baixo)
			min := new(big.Int)
			rangeMultiplier.Int(min)

			// Adicionando rangeMultiplier ao valor mínimo (privKeyMinInt)
			min.Add(privKeyMinInt, min)

			// Verificando o valor final como uma string hexadecimal
			verificaKey := min.Text(16)
			privKeyInt.SetString(verificaKey, 16)
			fmt.Printf("Porcentagem informada, iniciando: %s\n", verificaKey)

			}else{
				verificaKey = ranges.Ranges[rangeNumber-1].Min
				privKeyInt.SetString(verificaKey[2:], 16)
				fmt.Printf("Iniciando do começo. %s: %s\n", carteirasalva, verificaKey)
			}

		} else {
			fmt.Printf("Encontrada chave no arquivo ultimaChavePorCarteira.txt pela carteira %s: %s\n", carteirasalva, verificaKey)
			privKeyInt.SetString(verificaKey, 16)
		}
	}else{ //?

	}


	// Load wallet addresses from JSON file
	wallets, err := LoadWallets(filepath.Join(rootDir, "data", "wallets.json"))
	if err != nil {
		log.Fatalf("Failed to load wallets: %v", err)
	}

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
				if(modoSelecionado == 2){saveUltimaKeyWallet("ultimaChavePorCarteira.txt", carteirasalva, lastkey)}			
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

	if(modoSelecionado == 2){
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

