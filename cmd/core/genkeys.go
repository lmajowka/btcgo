/**
 * BTCGO
 *
 * Modulo : Generate Keys
 */

package core

import (
	"context"
	"math/big"
)

type GenKeys struct {
	// Context
	Ctx       context.Context
	CtxCancel context.CancelFunc
	// Channels

	// Data
	TotalGenKeys float64
	KeyChannel   chan *big.Int
	PrivKeyInt   *big.Int
	IsStarted    bool
}

// Criar Instancia
func NewGenKeys(ctx context.Context, keyChannel chan *big.Int) *GenKeys {
	newCtx, newCancel := context.WithCancel(ctx)
	return &GenKeys{
		Ctx:          newCtx,
		CtxCancel:    newCancel,
		TotalGenKeys: 0,
		KeyChannel:   keyChannel,
		PrivKeyInt:   big.NewInt(0),
		IsStarted:    false,
	}
}

// Start
func (g *GenKeys) Start() {
	go func() {
		for {
			privKeyCopy := new(big.Int).Set(g.PrivKeyInt)
			select {
			case g.KeyChannel <- privKeyCopy:
				g.PrivKeyInt.Add(g.PrivKeyInt, big.NewInt(1))
				g.TotalGenKeys++

			case <-g.Ctx.Done():
				return
			}
		}
	}()
}

// Get Total Gen Keys
func (g GenKeys) GetTotalKeys() float64 {
	return g.TotalGenKeys
}

// Get Last Key
func (g *GenKeys) GetLastKey() *big.Int {
	return g.PrivKeyInt
}

// Stop
func (g *GenKeys) Stop() {
	g.IsStarted = false
	if g.IsStarted {
		g.CtxCancel()
	}
}
