package utils

import (
	"btcgo/src/crypto/base58"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"fmt"
	"strings"
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

// Salva a ultima chave da carteira pesquisada.
func SaveUltimaKeyWallet(filename string, carteira string, chave string) error {
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