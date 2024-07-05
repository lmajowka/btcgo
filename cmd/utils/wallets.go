/**
 * BTCGO
 *
 * Modulo : Carteiras
 */

package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

// Wallet Struct
type Wallet struct {
	FileName         string
	SearchingWallets string
	DataWallet       map[string]bool
	DataWalletID     map[int]string
}

// Cria uma instancia
func NewWalletData(filename string) *Wallet {
	return &Wallet{
		FileName:         filename,
		SearchingWallets: "",
		DataWallet:       make(map[string]bool),
		DataWalletID:     make(map[int]string),
	}
}

// Ler Carteiras para memoria
func (w *Wallet) Load() error {
	bytes, err := os.ReadFile(w.FileName)
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
	for i, address := range walletsTemp.Addresses {
		w.DataWallet[address] = true
		w.DataWalletID[i] = address
	}
	return nil
}

// Verifica se a carteira Existe
func (w Wallet) Exist(wallet string) bool {
	if _, ok := w.DataWallet[wallet]; ok {
		return true
	}
	return false
}

// Set Wallet to Find
func (w *Wallet) SetFindWallet(walletid int) {
	w.SearchingWallets = w.DataWalletID[walletid-1]
	fmt.Println(w.DataWalletID[walletid-1])
}

// Is this Wallet need Find
func (w *Wallet) IsSearchWallet(wallet string) bool {
	return w.SearchingWallets == wallet
}
