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
	"github.com/dustin/go-humanize"
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

func main() {
	green := color.New(color.FgGreen).SprintFunc()

	// Carregar ranges do arquivo JSON
	ranges, err := loadRanges("ranges.json")
	if err != nil {
		log.Fatalf("Falha ao carregar ranges: %v", err)
	}

	color.Cyan("BTCGO - Investidor Internacional")
	color.White("v0.1")

	// Perguntar ao usuário o número do range
	rangeNumber := promptRangeNumber(len(ranges.Ranges))

	// Inicializar privKeyInt com o valor máximo do range selecionado
	privKeyHex := ranges.Ranges[rangeNumber-1].Max

	privKeyInt := new(big.Int)
	privKeyInt.SetString(privKeyHex[2:], 16)

	// Carregar endereços de carteira do arquivo JSON
	wallets, err := loadWallets("wallets.json")
	if err != nil {
		log.Fatalf("Falha ao carregar carteiras: %v", err)
	}

	keysChecked := 0
	startTime := time.Now()

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
	case <-time.After(time.Minute * 100000000): // Opcional: Timeout após 10 minutos
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
}

func worker(wallets *Wallets, privKeyChan <-chan *big.Int, resultChan chan<- *big.Int, wg *sync.WaitGroup) {
	defer wg.Done()
	for privKeyInt := range privKeyChan {
		address := createPublicAddress(privKeyInt)
		//fmt.Printf("Endereço publico: ", address)
		if contains(wallets.Addresses, address) {
			resultChan <- privKeyInt
			return
		}
	}
}

func createPublicAddress(privKeyInt *big.Int) string {
	privKeyHex := fmt.Sprintf("%064x", privKeyInt)

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
}

// hash160 calcula o hash RIPEMD160(SHA256(b)).
func hash160(b []byte) []byte {
	h := sha256.New()
	h.Write(b)
	sha256Hash := h.Sum(nil)

	r := ripemd160.New()
	r.Write(sha256Hash)
	return r.Sum(nil)
}

// encodeAddress codifica o hash da chave pública em um endereço Bitcoin.
func encodeAddress(pubKeyHash []byte, params *chaincfg.Params) string {
	versionedPayload := append([]byte{params.PubKeyHashAddrID}, pubKeyHash...)
	checksum := doubleSha256(versionedPayload)[:4]
	fullPayload := append(versionedPayload, checksum...)
	return base58Encode(fullPayload)
}

// doubleSha256 calcula SHA256(SHA256(b)).
func doubleSha256(b []byte) []byte {
	first := sha256.Sum256(b)
	second := sha256.Sum256(first[:])
	return second[:]
}

// base58Encode codifica um slice de bytes em uma string codificada em base58.
var base58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

func base58Encode(input []byte) string {
	var result []byte
	x := new(big.Int).SetBytes(input)

	base := big.NewInt(int64(len(base58Alphabet)))
	zero := big.NewInt(0)
	mod := &big.Int{}

	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		result = append(result, base58Alphabet[mod.Int64()])
	}

	// Inverter o resultado
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	// Adicionar zeros à esquerda
	for _, b := range input {
		if b != 0 {
			break
		}
		result = append([]byte{base58Alphabet[0]}, result...)
	}

	return string(result)
}

// loadWallets carrega endereços de carteiras de um arquivo JSON
func loadWallets(filename string) (*Wallets, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var wallets Wallets
	if err := json.Unmarshal(bytes, &wallets); err != nil {
		return nil, err
	}

	return &wallets, nil
}

// contains verifica se uma string está em um slice de strings
func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

// loadRanges carrega ranges de um arquivo JSON
func loadRanges(filename string) (*Ranges, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var ranges Ranges
	if err := json.Unmarshal(bytes, &ranges); err != nil {
		return nil, err
	}

	return &ranges, nil
}

// promptRangeNumber solicita ao usuário que selecione um número de range
func promptRangeNumber(totalRanges int) int {
	reader := bufio.NewReader(os.Stdin)
	charReadline := '\n'

	if runtime.GOOS == "windows" {
		charReadline = '\r'
	}

	for {
		fmt.Printf("Escolha a carteira (1 a %d): ", totalRanges)
		input, _ := reader.ReadString(byte(charReadline))
		input = strings.TrimSpace(input)
		rangeNumber, err := strconv.Atoi(input)
		if err == nil && rangeNumber >= 1 && rangeNumber <= totalRanges {
			return rangeNumber
		}
		fmt.Println("Número inválido.")
	}
}

// calculatePercentage calcula a porcentagem da chave privada encontrada dentro do range
func calculatePercentage(privKeyInt, minKeyInt, maxKeyInt *big.Int) float64 {
	totalRange := new(big.Int).Sub(maxKeyInt, minKeyInt)
	foundPosition := new(big.Int).Sub(privKeyInt, minKeyInt)

	percentage := new(big.Float).Quo(
		new(big.Float).SetInt(foundPosition),
		new(big.Float).SetInt(totalRange),
	)
	percentage.Mul(percentage, big.NewFloat(100))

	percentageFloat, _ := percentage.Float64()
	return percentageFloat
}
