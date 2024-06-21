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

// Wallets struct to hold the array of wallet addresses
type Wallets struct {
	Addresses []string `json:"wallets"`
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

	ranges, err := loadRanges("ranges.json")
	if err != nil {
		log.Fatalf("Failed to load ranges: %v", err)
	}

	color.Cyan("BTCGO - Investidor Internacional")
	color.White("v0.1")

	// Ask the user for the range number
	rangeNumber := promptRangeNumber(len(ranges.Ranges))

	// Initialize privKeyInt with the minimum value of the selected range
	privKeyHex := ranges.Ranges[rangeNumber-1].Min

	privKeyInt := new(big.Int)
	privKeyInt.SetString(privKeyHex[2:], 16)

	// Load wallet addresses from JSON file
	wallets, err := loadWallets("wallets.json")
	if err != nil {
		log.Fatalf("Failed to load wallets: %v", err)
	}

	keysChecked := 0
	startTime := time.Now()

	// Number of CPU cores to use
	numCPU := runtime.NumCPU()
	fmt.Printf("CPUs detectados: %s\n", green(numCPU))
	runtime.GOMAXPROCS(numCPU * 2)

	// Create a channel to send private keys to workers
	privKeyChan := make(chan *big.Int)
	// Create a channel to receive results from workers
	resultChan := make(chan *big.Int)
	// Create a wait group to wait for all workers to finish
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < numCPU*2; i++ {
		wg.Add(1)
		go worker(wallets, privKeyChan, resultChan, &wg)
	}

	// Ticker for periodic updates every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	done := make(chan bool)

	// Goroutine to print speed updates
	go func() {
		for {
			select {
			case <-ticker.C:
				elapsedTime := time.Since(startTime).Seconds()
				keysPerSecond := float64(keysChecked) / elapsedTime
				fmt.Printf("Chaves checadas: %s, Chaves por segundo: %s\n", humanize.Comma(int64(keysChecked)), humanize.Comma(int64(keysPerSecond)))
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	// Send private keys to the workers
	go func() {
		for i := 1; i < 2; {
			privKeyCopy := new(big.Int).Set(privKeyInt)
			privKeyChan <- privKeyCopy
			privKeyInt.Add(privKeyInt, big.NewInt(1))
			keysChecked++
		}
		close(privKeyChan)
	}()
	// Wait for a result from any worker
	var foundAddress *big.Int
	select {
	case foundAddress = <-resultChan:
		color.Yellow("Chave privada encontrada: %064x\n", foundAddress)
		// close(resultChan)
	case <-time.After(time.Minute * 10): // Optional: Timeout after 1 minute
		fmt.Println("No address found within the time limit.")
	}

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(privKeyChan)
	}()

	elapsedTime := time.Since(startTime).Seconds()
	keysPerSecond := float64(keysChecked) / elapsedTime

	fmt.Printf("Chaves checadas: %s\n", humanize.Comma(int64(keysChecked)))
	fmt.Printf("Tempo: %.2f seconds\n", elapsedTime)
	fmt.Printf("Chaves por segundo: %s\n", humanize.Comma(int64(keysPerSecond)))
}

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

func createPublicAddress(privKeyInt *big.Int) string {

	privKeyHex := fmt.Sprintf("%064x", privKeyInt)

	// Decode the hexadecimal private key
	privKeyBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new private key using the secp256k1 package
	privKey := secp256k1.PrivKeyFromBytes(privKeyBytes)

	// Get the corresponding public key in compressed format
	compressedPubKey := privKey.PubKey().SerializeCompressed()

	// Generate a Bitcoin address from the public key
	pubKeyHash := hash160(compressedPubKey)
	address := encodeAddress(pubKeyHash, &chaincfg.MainNetParams)

	return address

}

// hash160 computes the RIPEMD160(SHA256(b)) hash.
func hash160(b []byte) []byte {
	h := sha256.New()
	h.Write(b)
	sha256Hash := h.Sum(nil)

	r := ripemd160.New()
	r.Write(sha256Hash)
	return r.Sum(nil)
}

// encodeAddress encodes the public key hash into a Bitcoin address.
func encodeAddress(pubKeyHash []byte, params *chaincfg.Params) string {
	versionedPayload := append([]byte{params.PubKeyHashAddrID}, pubKeyHash...)
	checksum := doubleSha256(versionedPayload)[:4]
	fullPayload := append(versionedPayload, checksum...)
	return base58Encode(fullPayload)
}

// doubleSha256 computes SHA256(SHA256(b)).
func doubleSha256(b []byte) []byte {
	first := sha256.Sum256(b)
	second := sha256.Sum256(first[:])
	return second[:]
}

// base58Encode encodes a byte slice to a base58-encoded string.
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

	// Reverse the result
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	// Add leading zeroes
	for _, b := range input {
		if b != 0 {
			break
		}
		result = append([]byte{base58Alphabet[0]}, result...)
	}

	return string(result)
}

// loadWallets loads wallet addresses from a JSON file
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

// contains checks if a string is in a slice of strings
func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

// loadRanges loads ranges from a JSON file
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

// promptRangeNumber prompts the user to select a range number
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
		fmt.Println("Numero invalido.")
	}
}
