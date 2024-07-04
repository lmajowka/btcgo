/**
 * BTCGO
 *
 * Modulo : Processa Ultima Chave
 */

package utils

import (
	"encoding/json"
	"errors"
	"os"
	"time"
)

// File Data
type FileDataStuct struct {
	Chave    string `json:"chave"`
	DataHora string `json:"datahora"`
}

// LastKeyWallet Struct
type LastKeyWallet struct {
	FileName string
	Data     map[string]FileDataStuct
}

// Cria uma instancia
func NewLastKeyWallet(filename string) *LastKeyWallet {
	return &LastKeyWallet{
		FileName: filename,
		Data:     make(map[string]FileDataStuct),
	}
}

// Ler fichiero para memoria
func (l *LastKeyWallet) Load() error {
	// Busca todo o conteúdo atual do arquivo
	data, err := os.ReadFile(l.FileName)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &l.Data)
	if err != nil {
		return err
	}
	return nil
}

// Save memory to file
func (l LastKeyWallet) Save() error {
	// salva no arquivo com as modificações
	xData, err := json.Marshal(l.Data)
	if err == nil {
		err = os.WriteFile(l.FileName, xData, 0644)
		if err == nil {
			return nil
		}
	}
	return err
}

// Get last key
func (l LastKeyWallet) GetLastKey(carteira string) (string, error) {
	if _, ok := l.Data[carteira]; ok {
		return l.Data[carteira].Chave, nil
	}
	return "", errors.New("not found")
}

// Load and get last Key
func (l LastKeyWallet) LoadGetLastKey(carteira string) (string, error) {
	err := l.Load()
	if err == nil {
		return l.GetLastKey(carteira)
	}
	return "", err
}

// Set last key
func (l *LastKeyWallet) SetLastKey(carteira string, key string) {
	l.Data[carteira] = FileDataStuct{
		Chave:    key,
		DataHora: time.Now().Format("2006-01-02 15:04:05"),
	}
}

// Set and save last Key
func (l LastKeyWallet) SetSaveLastKey(carteira string, key string) error {
	l.Data[carteira] = FileDataStuct{
		Chave:    key,
		DataHora: time.Now().Format("2006-01-02 15:04:05"),
	}
	return l.Save()
}
