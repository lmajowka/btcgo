package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/fatih/color"
	"golang.org/x/crypto/ripemd160"
)

// Wallets struct para armazenar o array de endereços de carteiras
type Wallets struct {
	Addresses []string `json:"wallets"`
}

// Range struct para armazenar o mínimo, máximo e status
type Range struct {
	Min    string `json:"min"`
	Max    string `json:"max"`
	Status int    `json:"status"`
}

// Ranges struct para armazenar um array de ranges
type Ranges struct {
	Ranges []Range `json:"ranges"`
}

// State struct para armazenar o estado do processo
type State struct {
	LastCheckedKey string `json:"lastCheckedKey"`
	KeysChecked    int    `json:"keysChecked"`
}

func main() {
	green := color.New(color.FgGreen).SprintFunc()

	// Carregar ranges do arquivo JSON
	ranges, err := loadRanges("ranges.json")
	if err != nil {
		log.Fatalf("Falha ao carregar ranges: %v", err)
	}

	color.Cyan("BTCGO - Investidor Internacional")
	color.White("v0.1")

	// Perguntar ao usuário se deseja continuar de onde parou
	var continueFromLast bool
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Deseja continuar de onde parou? (S/N): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		input = strings.TrimSpace(input)
		input = strings.ToUpper(input)
		if input == "S" {
			continueFromLast = true
			break
		} else if input == "N" {
			continueFromLast = false
			break
		} else {
			fmt.Println("Resposta inválida. Por favor, digite S para Sim ou N para Não.")
		}
	}

	var lastCheckedKey string
	var keysChecked int

	// Se o usuário optar por continuar de onde parou, carregue o estado anterior
	if continueFromLast {
		state, err := loadState("state.json")
		if err != nil {
			log.Fatalf("Falha ao carregar estado anterior: %v", err)
		}
		lastCheckedKey = state.LastCheckedKey
		keysChecked = state.KeysChecked
		fmt.Printf("Continuando de %s com %d chaves verificadas.\n", lastCheckedKey, keysChecked)
	} else {
		fmt.Println("Iniciando a verificação do início.")
	}

	// Número de núcleos de CPU a serem utilizados
	numCPU := runtime.NumCPU()
	fmt.Printf("CPUs detectados: %s\n", green(numCPU))
	runtime.GOMAXPROCS(numCPU * 2)

	// Criar um canal para enviar chaves privadas aos trabalhadores
	privKeyChan := make(chan *big.Int)
	// Criar um canal para receber resultados dos trabalhadores
	resultChan := make(chan *big.Int)
	// Criar um grupo de espera para aguardar todos os trabalhadores terminarem
	var wg sync.WaitGroup

	// Iniciar goroutines de trabalhadores
	for i := 0; i < numCPU*2; i++ {
		wg.Add(1)
		go worker(wallets, privKeyChan, resultChan, &wg)
	}

	// Ticker para atualizações periódicas a cada 5 segundos
	ticker := time.NewTicker(5 * time.Second)
	done := make(chan bool)

	// Definir o range total para cálculo de porcentagem
	minKeyInt := new(big.Int)
	minKeyInt.SetString(ranges.Ranges[rangeNumber-1].Min[2:], 16)
	maxKeyInt := new(big.Int)
	maxKeyInt.SetString(ranges.Ranges[rangeNumber-1].Max[2:], 16)
	totalKeys := new(big.Int).Sub(maxKeyInt, minKeyInt)

	// Goroutine para imprimir atualizações de velocidade
	go func() {
		for {
			select {
			case <-ticker.C:
				elapsedTime := time.Since(startTime).Seconds()
				keysPerSecond := float64(keysChecked) / elapsedTime
				checkedKeys := new(big.Int).Sub(maxKeyInt, privKeyInt)
				percentageChecked := new(big.Float).Quo(new(big.Float).SetInt(checkedKeys), new(big.Float).SetInt(totalKeys))
				percentageChecked.Mul(percentageChecked, big.NewFloat(100))
				percentageCheckedFloat, _ := percentageChecked.Float64()

				fmt.Printf("Chaves checadas: %s, Chaves por segundo: %s, Porcentagem checada: %.2f%%\n", humanize.Comma(int64(keysChecked)), humanize.Comma(int64(keysPerSecond)), percentageCheckedFloat)

			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	// Enviar chaves privadas aos trabalhadores
	go func() {
		for privKeyInt.Cmp(minKeyInt) >= 0 {
			privKeyCopy := new(big.Int).Set(privKeyInt)
			privKeyChan <- privKeyCopy
			privKeyInt.Sub(privKeyInt, big.NewInt(1))
			keysChecked++
		}
		close(privKeyChan)
	}()

	// Aguardar um resultado de qualquer trabalhador
	var foundAddress *big.Int
	select {
	case foundAddress = <-resultChan:
		color.Yellow("Chave privada encontrada: %064x\n", foundAddress)

		// Chave privada encontrada, formatando a saída
		addressInfo := fmt.Sprintf("Chave privada encontrada: %064x\n", foundAddress)

		// Calculando a porcentagem
		percentage := calculatePercentage(foundAddress, minKeyInt, maxKeyInt)
		fmt.Printf("Porcentagem da chave encontrada: %.2f%%\n", percentage)

		// Criando um arquivo para registrar a chave encontrada
		fileName := "Chave_encontrada.txt"
		file, err := os.Create(fileName)
		if err != nil {
			fmt.Println("Erro ao criar o arquivo:", err)
			return
		}
		defer file.Close()

		// Escrevendo a informação no arquivo
		_, err = file.WriteString(fmt.Sprintf("%s\nPorcentagem: %.2f%%\n", addressInfo, percentage))
		if err != nil {
			fmt.Println("Erro ao escrever no arquivo:", err)
			return
		}

		// Confirmação para o usuário
		color.Yellow(addressInfo)
		fmt.Printf("Chave privada encontrada e registrada em %s\n", fileName)
	case <-time.After(time.Minute * 10): // Opcional: Timeout após 10 minutos
		fmt.Println("Nenhum endereço encontrado dentro do limite de tempo.")
	}

	// Aguardar todos os trabalhadores terminarem
	go func() {
		wg.Wait()
		close(done)
	}()

	elapsedTime := time.Since(startTime).Seconds()
	keysPerSecond := float64(keysChecked) / elapsedTime
	checkedKeys := new(big.Int).Sub(maxKeyInt, privKeyInt)
	percentageChecked := new(big.Float).Quo(new(big.Float).SetInt(checkedKeys), new(big.Float).SetInt(totalKeys))
	percentageChecked.Mul(percentageChecked, big.NewFloat(100))
	percentageCheckedFloat, _ := percentageChecked.Float64()

	fmt.Printf("Chaves checadas: %s\n", humanize.Comma(int64(keysChecked)))
	fmt.Printf("Tempo: %.2f segundos\n", elapsedTime)
	fmt.Printf("Chaves por segundo: %s\n", humanize.Comma(int64(keysPerSecond)))
	fmt.Printf("Porcentagem checada: %.2f%%\n", percentageCheckedFloat)

	// Salvar o estado atual do processo antes de encerrar
	if foundAddress != nil {
		lastCheckedKey := fmt.Sprintf("%064x", foundAddress)
		if err := saveState("state.json", lastCheckedKey, keysChecked); err != nil {
			fmt.Printf("Erro ao salvar o estado atual: %v\n", err)
		}
	}
}

// worker é a função que processa cada chave privada
func worker(wallets *Wallets, privKeyChan <-chan *big.Int, resultChan chan<- *big.Int, wg *sync.WaitGroup) {
	defer wg.Done()
	for privKeyInt := range privKeyChan {
		address := createPublicAddress(privKeyInt)
		if contains(wallets.Addresses, address) {
			resultChan <- privKeyInt
			return
		}
	}
}

// createPublicAddress cria um endereço público a partir de uma chave privada
func createPublicAddress(privKeyInt *big.Int) string {
	privKey
	Hex := fmt.Sprintf("%064x", privKeyInt)
// Decodificar a chave privada hexadecimal
privKeyBytes, err := hex.DecodeString(privKeyHex)
if err != nil {
	log.Fatal(err)
}

// Criar uma nova chave privada usando o pacote secp256k1
privKey := secp256k1.PrivKeyFromBytes(privKeyBytes)

// Obter a chave pública correspondente no formato comprimido
compressedPubKey := privKey.PubKey().SerializeCompressed()

// Gerar um endereço Bitcoin a partir da chave pública
pubKeyHash := hash160(compressedPubKey)
address := encodeAddress(pubKeyHash, &chaincfg.MainNetParams)

return address
