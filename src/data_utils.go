package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Cria uma copia do ficheiro em memoria
type controlFileDataStuct struct {
	Chave    string `json:"chave"`
	DataHora string `json:"datahora"`
}

var controlFileData = make(map[string]controlFileDataStuct)

// loadRanges loads ranges from a JSON file
func LoadRanges(filename string) (*Ranges, error) {
	bytes, err := os.ReadFile(filename)
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
func LoadWallets(filename string, walletNumber int) error {
	bytes, err := os.ReadFile(filename)
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
	for i, address := range walletsTemp.Addresses {
		if (walletNumber - 1) == i {
			WallettoFind = address
		}
		Wallets[address] = address
	}
	return nil
}

// Save Control File
func saveUltimaKeyWallet(filename string, carteira string, chave string) error {
	// Actualiza dados
	if _, ok := controlFileData[carteira]; ok {
		xTmp := controlFileData[carteira]
		xTmp.Chave = chave
		xTmp.DataHora = time.Now().Format("2006-01-02 15:04:05")
		controlFileData[carteira] = xTmp
	} else {
		controlFileData[carteira] = controlFileDataStuct{
			Chave:    chave,
			DataHora: time.Now().Format("2006-01-02 15:04:05"),
		}
	}
	// salva no arquivo com as modificações
	xData, err := json.Marshal(controlFileData)
	if err == nil {
		err = os.WriteFile(filename, xData, 0644)
		if err == nil {
			return nil
		}
	}
	return err
}

// Load Control File
func LoadUltimaKeyWallet(filename string, carteira string) (string, error) {
	// Busca todo o conteúdo atual do arquivo
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(data, &controlFileData)
	if err == nil {
		// Verifica se existe a carteira
		if _, ok := controlFileData[carteira]; ok {
			return controlFileData[carteira].Chave, nil
		}
	}
	return "", fmt.Errorf("carteira %s não encontrada", carteira)
}
