/**
 * BTCGO
 *
 * Modulo : Workers
 */

package core

import (
	"btcgo/cmd/utils"
	"context"
	"math/big"
	"sync"
	"time"
)

type workerStruct struct {
}

type Workers struct {
	// Context
	Ctx       context.Context
	CtxCancel context.CancelFunc

	// Channels
	ResultChannel chan *utils.ResultDataStruct
	KeyChannel    chan *big.Int
	Wg            sync.WaitGroup

	// Data
	worker    []workerStruct
	IsStarted bool
}

// Criar Instancia
func NewWorkers(ctx context.Context, keych chan *big.Int, resultChannel chan *utils.ResultDataStruct) *Workers {
	newCtx, newCancel := context.WithCancel(ctx)
	return &Workers{
		Ctx:           newCtx,
		CtxCancel:     newCancel,
		worker:        []workerStruct{},
		KeyChannel:    keych,
		ResultChannel: resultChannel,
		IsStarted:     false,
	}
}

// Start
func (w *Workers) Start() {
	for i := 0; i < App.MaxWorkers; i++ {
		w.Wg.Add(1)
		go func() {
			defer w.Wg.Done()
			w.run()
		}()
	}
	w.IsStarted = true

	<-w.Ctx.Done()
	w.IsStarted = false
}

// Start Workers
func (w *Workers) run() {
	for privKeyInt := range w.KeyChannel {
		if w.Ctx.Err() != nil { // StopRequest
			return
		}
		address := utils.CreatePublicHash160(privKeyInt)
		wallet := utils.Hash160ToAddress(address)

		// Verificar se a chave está na carteira
		if App.Wallets.Exist(wallet) {
			w.ResultChannel <- &utils.ResultDataStruct{
				Wallet:   wallet,
				Key:      privKeyInt,
				Wif:      utils.GenerateWif(privKeyInt),
				HoraData: time.Now().Format("2006-01-02 15:04:05"),
			}
			// Verifica se era esta a carteira que procurava
			if App.Wallets.IsSearchWallet(wallet) {
				App.Stop(true)
				return
			}
		}
	}
}

// Stop Workers
func (w *Workers) Stop() {
	if w.IsStarted {
		w.CtxCancel()
	}
}
