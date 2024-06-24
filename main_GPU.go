package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"golang.org/x/crypto/ripemd160"
)

type Wallets struct {
	Addresses []string `json:"wallets"`
}

func main() {
	green := color.New(color.FgGreen).SprintFunc()

	// Carregar endereços de carteira do arquivo JSON
	wallets, err := loadWallets("wallets.json")
	if err != nil {
		log.Fatalf("Falha ao carregar carteiras: %v", err)
	}

	// Carregar chaves privadas geradas do arquivo JSON
	keys, err := loadGeneratedKeys("keys.json")
	if err != nil {
		log.Fatalf("Falha ao carregar chaves geradas: %v", err)
	}

	keysChecked := 0
	startTime := time.Now()

	// Número de núcleos de CPU a serem utilizados
	numCPU := runtime.NumCPU()
	fmt.Printf("CPUs detectados: %s\n", green(numCPU))
	runtime.GOMAXPROCS(numCPU * 2)

	var wg sync.WaitGroup

	for _, key := range keys {
		wg.Add(1)
		go func(privKeyInt *big.Int) {
			defer wg.Done()
			address := createPublicAddress(privKeyInt)
			if contains(wallets.Addresses, address) {
				color.Yellow("Chave privada encontrada: %064x\n", privKeyInt)

				// Chave privada encontrada, formatando a saída
				addressInfo := fmt.Sprintf("Chave privada encontrada: %064x\n", privKeyInt)

				// Criando um arquivo para registrar a chave encontrada
				fileName := "Chave_encontrada.txt"
				file, err := os.Create(fileName)
				if err != nil {
					fmt.Println("Erro ao criar o arquivo:", err)
					return
				}
				defer file.Close()

				// Escrevendo a informação no arquivo
				_, err = file.WriteString(addressInfo)
				if err != nil {
					fmt.Println("Erro ao escrever no arquivo:", err)
					return
				}

				// Confirmação para o usuário
				color.Yellow(addressInfo)
				fmt.Println("Chave privada encontrada e registrada em", fileName)
			}
			keysChecked++
		}(key)
	}

	wg.Wait()

	elapsedTime := time.Since(startTime).Seconds()
	keysPerSecond := float64(keysChecked) / elapsedTime

	fmt.Printf("Chaves checadas: %s\n", humanize.Comma(int64(keysChecked)))
	fmt.Printf("Tempo: %.2f segundos\n", elapsedTime)
	fmt.Printf("Chaves por segundo: %s\n", humanize.Comma(int64(keysPerSecond)))
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

// contains verifica se uma string está em um slice de strings
func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

// loadGeneratedKeys carrega chaves geradas de um arquivo JSON
func loadGeneratedKeys(filename string) ([]*big.Int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var keysList []string
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bytes, &keysList); err != nil {
		return nil, err
	}

	keys := make([]*big.Int, len(keysList))
	for i, key := range keysList {
		keyInt := new(big.Int)
		keyInt.SetString(strings.TrimPrefix(key, "0x"), 16)
		keys[i] = keyInt
	}

	return keys, nil
}
