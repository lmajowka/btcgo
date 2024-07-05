/**
 * BTCGO
 *
 * Modulo : Gravar Resultados
 */

package utils

import (
	"context"
	"fmt"
	"math/big"
	"os"

	"github.com/fatih/color"
)

type ResultDataStruct struct {
	Wallet   string
	Key      big.Int
	Wif      string
	HoraData string
}

type Results struct {
	// Context
	Ctx    context.Context
	Cancel context.CancelFunc
	// Channels
	ResultChannel chan *ResultDataStruct
	// Data
	FileName  string
	IsStarted bool
}

// Cria uma instancia
func NewResults(mainctx context.Context, resultChannel chan *ResultDataStruct, filename string) *Results {
	newContext, NewCancel := context.WithCancel(mainctx)
	return &Results{
		FileName:      filename,
		Ctx:           newContext,
		Cancel:        NewCancel,
		ResultChannel: resultChannel,
		IsStarted:     false,
	}
}

// Save Data
func (rs Results) saveData(data *ResultDataStruct) error {
	file, err := os.OpenFile(rs.FileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer file.Close()
		_, err = file.WriteString(fmt.Sprintf("Data/Hora: %s | Chave privada: %064x | WIF: %s | Wallet: %s\n", data.HoraData, &data.Key, data.Wif, data.Wallet))
	}
	return err
}

// Start
func (rs *Results) Start() {
	go func() {
		for {
			select {
			case data := <-rs.ResultChannel:
				color.Yellow("Wallet: %s\n", data.Wallet)
				color.Yellow("Chave privada encontrada: %064x\n", &data.Key)
				color.Yellow("WIF: %s", data.Wif)
				if err := rs.saveData(data); err != nil {
					fmt.Println("Erro ao escrever no arquivo:", err)
				}

			case <-rs.Ctx.Done():
				return
			}
		}
	}()
	rs.IsStarted = true
}

// Stop
func (rs *Results) Stop() {
	rs.IsStarted = false
	if rs.IsStarted {
		rs.Cancel()
	}
}
