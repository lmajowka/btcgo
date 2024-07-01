package main

import (
	"btcgo/src/crypto/base58"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

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
func LoadWallets(filename string) error {
	file, err := os.Open(filename)
	if err != nil {

		return err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	type WalletsTemp struct {
		Addresses []string `json:"wallets"`
	}
	var walletsTemp WalletsTemp
	if err := json.Unmarshal(bytes, &walletsTemp); err != nil {
		return err
	}
	//var wallets Wallets
	for _, address := range walletsTemp.Addresses {
		Wallets[string(base58.Decode(address)[1:21])] = base58.Decode(address)[1:21]
	}
	return nil
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
