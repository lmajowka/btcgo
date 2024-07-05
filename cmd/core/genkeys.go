/**
 * BTCGO
 *
 * Modulo : Generate Keys
 */

package core

import (
	"context"
	"fmt"
	"math/big"
	"math/rand/v2"
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

// Find Start Position
func (g *GenKeys) findStartPos() {
	switch App.Modo {
	case 1:
		g.posInicio()

	case 2:
		if !App.DesdeInicio {
			key, err := App.LastKey.GetLastKey(App.Carteira)
			if err != nil || key == "" {
				g.fromPercent()
				fmt.Printf("Range informado, iniciando: %s\n", key)
			} else {
				g.PrivKeyInt.SetString(key, 16)
				fmt.Printf("Encontrada chave no arquivo. Carteira %s: %s\n", App.Carteira, key)
			}
		} else {
			g.posInicio()
		}

	case 3:
		App.StartPosPercent = g.genRandom()
		g.fromPercent()
	}
}

// Procura o inicio/random atravez de %
func (g *GenKeys) fromPercent() {
	privKeyMinInt := new(big.Int)
	privKeyMaxInt := new(big.Int)
	privKeyMin, _ := App.Ranges.GetMin(App.RangeNumber)
	privKeyMax, _ := App.Ranges.GetMax(App.RangeNumber)
	privKeyMinInt.SetString(privKeyMin[2:], 16)
	privKeyMaxInt.SetString(privKeyMax[2:], 16)
	// Calculando a diferença entre privKeyMaxInt e privKeyMinInt
	rangeKey := new(big.Int).Sub(privKeyMaxInt, privKeyMinInt)
	// Calculando o valor de rangeKey multiplicado pela porcentagem
	rangeMultiplier := new(big.Float).Mul(new(big.Float).SetInt(rangeKey), big.NewFloat(App.StartPosPercent/100.0))
	// Convertendo o resultado para inteiro (arredondamento para baixo)
	min := new(big.Int)
	rangeMultiplier.Int(min)
	// Adicionando rangeMultiplier ao valor mínimo (privKeyMinInt)
	min.Add(privKeyMinInt, min)
	// Verificando o valor final como uma string hexadecimal
	key := min.Text(16)
	g.PrivKeyInt.SetString(key, 16)
}

// Set o Range desde o inicio
func (g *GenKeys) posInicio() {
	privKeyHex, _ := App.Ranges.GetMin(App.RangeNumber)
	//log.Println(privKeyHex, a.RangeNumber)
	g.PrivKeyInt.SetString(privKeyHex[2:], 16)
}

// Start
func (g *GenKeys) Start() {
	// Find Start Position
	g.findStartPos()
	go func() {
		xTmpRandomCtrl := 0 // Despois de fazer random testa 10000 chaves seguintes
		for {
			privKeyCopy := new(big.Int).Set(g.PrivKeyInt)
			select {
			case g.KeyChannel <- privKeyCopy:
				if App.Modo == 3 {
					App.Modo = 31
					App.StartPosPercent = g.genRandom()
					g.fromPercent()
				} else {
					g.PrivKeyInt.Add(g.PrivKeyInt, big.NewInt(1))
					if App.Modo == 31 {
						if xTmpRandomCtrl > 10000 {
							App.Modo = 3
							xTmpRandomCtrl = 0
						}
						xTmpRandomCtrl++
					}
				}
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

// Gerar um valor % random
func (g *GenKeys) genRandom() float64 {
	min := float64(0)
	max := float64(100)
	ranF := (min + rand.Float64()*(max-min))
	return ranF
}

// Stop
func (g *GenKeys) Stop() {
	g.IsStarted = false
	if g.IsStarted {
		g.CtxCancel()
	}
}
