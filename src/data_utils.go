package main

import (
	"btcgo/src/crypto/base58"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
<<<<<<< Updated upstream
)

=======
	"runtime"
	"strconv"
	"strings"
	"math/big"
	"sync"

	"btcgo/src/crypto/btc_utils"
)

// promptRangeNumber prompts the user to select a range number
func PromptRangeNumber(totalRanges int) int {
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

// PromptModos prompts the user to select a modo's
func PromptModos(totalModos int) int {
	reader := bufio.NewReader(os.Stdin)
	charReadline := '\n'

	if runtime.GOOS == "windows" {
		charReadline = '\r'
	}

	for {
		fmt.Printf("Escolha os modos que deseja de (1 a %d) \n  Modo do inicio: 1 - Modo sequencial(chave do arquivo): 2): ", totalModos)
		input, _ := reader.ReadString(byte(charReadline))
		input = strings.TrimSpace(input)
		modoSelecinado, err := strconv.Atoi(input)
		if err == nil && modoSelecinado >= 1 && modoSelecinado <= totalModos {
			return modoSelecinado
			//fmt.Println(modoSelecinado)
		}
		fmt.Println("Modo invalido.")
	}
}

// PromptAuto solicita ao usuário a seleção de um número dentro de um intervalo específico.
func PromptAuto(pergunta string, totalnumbers int) int {
	reader := bufio.NewReader(os.Stdin)
	charReadline := '\n'

	if runtime.GOOS == "windows" {
		charReadline = '\r'
	}

	for {
		fmt.Printf(pergunta)
		input, _ := reader.ReadString(byte(charReadline))
		input = strings.TrimSpace(input)
		resposta, err := strconv.Atoi(input)
		if err == nil && resposta >= 1 && resposta <= totalnumbers {
			return resposta
		}
		fmt.Println("Resposta inválida.")
	}
}


>>>>>>> Stashed changes
// contains checks if a string is in a slice of strings
func Contains(slice [][]byte, item []byte) bool {
	for _, a := range slice {
		if bytes.Equal(a, item) {
			return true
		}
	}
	return false
}

// loadRanges loads ranges from a JSON file
func LoadRanges(filename string) (*Ranges, error) {
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

// loadWallets loads wallet addresses from a JSON file
func LoadWallets(filename string) (*Wallets, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	type WalletsTemp struct {
		Addresses []string `json:"wallets"`
	}

	var walletsTemp WalletsTemp
	if err := json.Unmarshal(bytes, &walletsTemp); err != nil {
		return nil, err
	}

	var wallets Wallets
	for _, address := range walletsTemp.Addresses {
		wallets.Addresses = append(wallets.Addresses, base58.Decode(address)[1:21])
	}

	return &wallets, nil
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

func saveUltimaKeyWallet(filename string, carteira string, chave string) error {
	// abre o arquivo em modo de append, cria se não existir
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// busca todo o conteúdo atual do arquivo
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	found := false

	// verifica se a carteira já existe no arquivo
	for i, line := range lines {
		parts := strings.SplitN(line, "|", 2) // divide a linha em duas partes pelo " | "
		if len(parts) != 2 {
			continue // ignora linhas mal formatadas sem " | "
		}
		if parts[0] == carteira {
			// Substitui apenas a chave correspondente à carteira encontrada
			lines[i] = fmt.Sprintf("%s|%s", carteira, chave)
			found = true
			break
		}
	}
	// se a carteira não foi encontrada, adiciona uma nova linha
	if !found {
		lines = append(lines, fmt.Sprintf("%s|%s", carteira, chave))
	}
	// salva no arquivo com as modificações
	err = os.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		return err
	}

	return nil
}



func LoadUltimaKeyWallet(filename string, carteira string) (string, error) {
    // Busca todo o conteúdo atual do arquivo
    data, err := os.ReadFile(filename)
    if err != nil {
        return "", err
    }

    lines := strings.Split(string(data), "\n")

    // Verifica se a carteira já existe no arquivo
    for _, line := range lines {
        parts := strings.SplitN(line, "|", 2) // Divide a linha em duas partes pelo "|"
        if len(parts) != 2 {
            continue // Ignora linhas mal formatadas sem "|"
        }
        if parts[0] == carteira {
            return parts[1], nil
        }
    }

    // Retorna erro se a carteira não for encontrada
    return "", fmt.Errorf("carteira %s não encontrada", carteira)
}