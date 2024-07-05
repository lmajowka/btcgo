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

	NumRecsRandom int
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
		App.Modo = 31
		//log.Println("Encotrada uma nova chave", App.StartPosPercent, "%")
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
			if App.USEDB == 1 {
				if App.Modo == 3 || App.Modo == 31 {
					xFail := 0
					for {
						if App.Modo == 3 {
							if !App.DB.ExistKey(fmt.Sprintf("%064x", g.PrivKeyInt)) {
								break
							} else {
								//log.Println("(3)key tested", fmt.Sprintf("%064x", g.PrivKeyInt))
								App.StartPosPercent = g.genRandom()
								g.fromPercent()
								//log.Println("(3)find other", fmt.Sprintf("%064x", g.PrivKeyInt))
							}
						} else if App.Modo == 31 {
							if !App.DB.ExistKey(fmt.Sprintf("%064x", g.PrivKeyInt)) {
								break
							} else {
								if xFail > 100 {
									App.Modo = 3
									xFail = 0
								} else {
									//log.Println("(31)key tested", fmt.Sprintf("%064x", g.PrivKeyInt))
									x := g.PrivKeyInt.Add(g.PrivKeyInt, big.NewInt(1))
									g.PrivKeyInt = new(big.Int).Set(x)
									//log.Println("(31)find other", fmt.Sprintf("%064x \n %064x", x, g.PrivKeyInt))
									xFail++
								}
							}
						}
					}
					App.DB.InsertKey(App.Carteira, fmt.Sprintf("%064x", g.PrivKeyInt))
				}
			}

			privKeyCopy := new(big.Int).Set(g.PrivKeyInt)
			select {
			case g.KeyChannel <- privKeyCopy:
				if App.Modo == 3 {
					App.Modo = 31
					App.StartPosPercent = g.genRandom()
					g.fromPercent()
				} else {
					if App.Modo == 31 {
						if xTmpRandomCtrl > g.NumRecsRandom {
							// Na proxima chave volta a fazer um random
							App.Modo = 3
							xTmpRandomCtrl = 0
						} else {
							xTmpRandomCtrl++
						}
					}
					g.PrivKeyInt.Add(g.PrivKeyInt, big.NewInt(1))
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

// Set Numero de Resc no Random
func (g *GenKeys) SetRecs(recs int) {
	g.NumRecsRandom = recs
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
