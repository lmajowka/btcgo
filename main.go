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

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"golang.org/x/crypto/ripemd160"

	"gorgonia.org/gocudnn"
	"gorgonia.org/gorgonia"
	"gorgonia.org/tensor"
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
	// Initialize CUDA device
	if err := gocudnn.Initialize(); err != nil {
		log.Fatalf("Could not initialize CUDA: %v", err)
	}
	defer gocudnn.Shutdown()

	// Example of using Gorgonia for tensor operations on GPU
	g := gorgonia.NewGraph()

	// Create a tensor with some values
	backing := []float32{1, 2, 3, 4}
	a := tensor.New(tensor.Of(tensor.Float32), tensor.WithShape(2, 2), tensor.WithBacking(backing))

	// Transfer the tensor to GPU
	tensorGPU := gorgonia.NewTensor(g, a.Dtype(), a.Dims(), gorgonia.WithShape(a.Shape()...), gorgonia.WithValue(a), gorgonia.WithDevice(gorgonia.UseGPU()))

	// Perform some operations
	result := gorgonia.Must(gorgonia.Add(tensorGPU, tensorGPU))

	// Create a VM and run the graph
	vmachine := gorgonia.NewTapeMachine(g, gorgonia.BindDualValues(tensorGPU))
	defer vmachine.Close()

	if err := vmachine.RunAll(); err != nil {
		log.Fatalf("Failed to run the graph: %v", err)
	}

	// Extract the result
	fmt.Println("Result: ", result.Value())
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
