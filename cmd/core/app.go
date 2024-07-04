/**
 *	App BTCGO
 */

package core

import (
	"btcgo/cmd/utils"
	"context"
	"fmt"
	"log"
	"math/big"
	"path/filepath"
	"runtime"
	"time"
)

type AppStruct struct {
	Ctx       context.Context
	CtxCancel context.CancelFunc

	LastKey *utils.LastKeyWallet
	Wallets *utils.Wallet
	Ranges  *utils.Range
	Results *utils.Results
	Ticker  *TimerUpdater
	Workers *Workers
	Keys    *GenKeys

	// Channels
	ResultChannel chan *utils.ResultDataStruct
	KeyChannel    chan *big.Int

	// Data
	Carteira        string
	RangeNumber     int // id range slice
	Modo            int
	MaxWorkers      int
	DesdeInicio     bool
	StartPosPercent float64
}

var App *AppStruct

func NewApp() {
	// Create App Instance
	App = appInit()

	defer func() {
		close(App.ResultChannel)
		close(App.KeyChannel)
	}()

	// Load Files
	err := App.loadData()
	if err != nil {
		log.Fatalln(err)
	}

	// Start WebServer Api

	// Request Prompts
	App.consolePrompts()

	// Set Number Max of CPUs
	App.setCPUs()

	// Find Start Position
	App.findStartPos()

	// Start
	App.start()

	// Control CTRL+C
	/*
		SysCtrlSignal := make(chan os.Signal, 1)
		signal.Notify(SysCtrlSignal, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		for {
			select {
			case <-SysCtrlSignal:
				log.Println("Stop Request")
				App.Stop(true)
				return
			case <-App.Ctx.Done():
				log.Println("finished")
				App.Stop(false)
				return
			}
		}
	*/

	<-App.Ctx.Done()
	log.Println("finished")
}

// Init Application
func appInit() *AppStruct {
	rootDir, err := utils.GetPath()
	if err != nil {
		log.Panicln("Erro ao obter o caminho do executável")
	}

	newContext, newCancel := context.WithCancel(context.Background())
	var resultChannel = make(chan *utils.ResultDataStruct, 1)
	var keych = make(chan *big.Int)

	return &AppStruct{
		// Context
		Ctx:       newContext,
		CtxCancel: newCancel,
		// Channels
		ResultChannel: resultChannel,
		KeyChannel:    keych,

		// create instances
		LastKey: utils.NewLastKeyWallet(filepath.Join(rootDir, "data", "lastkeys.json")),
		Wallets: utils.NewWalletData(filepath.Join(rootDir, "data", "wallets.json")),
		Ranges:  utils.NewRanges(filepath.Join(rootDir, "data", "ranges.json")),
		Results: utils.NewResults(newContext, resultChannel, filepath.Join(rootDir, "chaves_encontradas.txt")),
		Ticker:  NewTicker(newContext),
		Workers: NewWorkers(newContext, keych, resultChannel),
		Keys:    NewGenKeys(newContext, keych),
	}
}

// Loading data
func (a *AppStruct) loadData() error {
	// Ultimas Chaves Processadas
	a.LastKey.Load()
	// Carregar as carteiras em memoria
	if err := a.Wallets.Load(); err != nil {
		log.Println("Não foi possivel ler as carteiras")
		return err
	}
	// Carrega Ranges
	if err := a.Ranges.Load(); err != nil {
		log.Println("Não foi possivel ler oa ranges.")
		return err
	}
	return nil
}

// Get User request
func (a *AppStruct) consolePrompts() {
	RequestData()
}

// Set CPUs
func (a *AppStruct) setCPUs() {
	runtime.GOMAXPROCS(a.MaxWorkers)
}

// Find Start Position
func (a *AppStruct) findStartPos() {
	switch a.Modo {
	case 1:
		a.posInicio()

	case 2:
		if !a.DesdeInicio {
			key, err := a.LastKey.GetLastKey(a.Carteira)
			if err != nil || key == "" {
				privKeyMinInt := new(big.Int)
				privKeyMaxInt := new(big.Int)
				privKeyMin, _ := a.Ranges.GetMin(a.RangeNumber)
				privKeyMax, _ := a.Ranges.GetMax(a.RangeNumber)
				privKeyMinInt.SetString(privKeyMin[2:], 16)
				privKeyMaxInt.SetString(privKeyMax[2:], 16)
				// Calculando a diferença entre privKeyMaxInt e privKeyMinInt
				rangeKey := new(big.Int).Sub(privKeyMaxInt, privKeyMinInt)
				// Calculando o valor de rangeKey multiplicado pela porcentagem
				rangeMultiplier := new(big.Float).Mul(new(big.Float).SetInt(rangeKey), big.NewFloat(a.StartPosPercent/100.0))
				// Convertendo o resultado para inteiro (arredondamento para baixo)
				min := new(big.Int)
				rangeMultiplier.Int(min)
				// Adicionando rangeMultiplier ao valor mínimo (privKeyMinInt)
				min.Add(privKeyMinInt, min)
				// Verificando o valor final como uma string hexadecimal
				key := min.Text(16)
				a.Keys.PrivKeyInt.SetString(key, 16)
				fmt.Printf("Range informado, iniciando: %s\n", key)
			} else {
				a.Keys.PrivKeyInt.SetString(key, 16)
				fmt.Printf("Encontrada chave no arquivo. Carteira %s: %s\n", a.Carteira, key)
			}
		} else {
			a.posInicio()
		}
	}
}

// Set o Range desde o inicio
func (a *AppStruct) posInicio() {
	privKeyHex, _ := a.Ranges.GetMin(a.RangeNumber)
	//log.Println(privKeyHex, a.RangeNumber)
	a.Keys.PrivKeyInt.SetString(privKeyHex[2:], 16)
}

// Start App Calc
func (a *AppStruct) start() {
	// Change Channel Size
	a.KeyChannel = make(chan *big.Int, a.MaxWorkers)
	// Start
	a.Results.Start() // Start Rotina que grava os resultados
	a.Ticker.Start(5) // Inicia as actualizaçóes da ultima chave
	a.Keys.Start()    // Gerar Chaves
	a.Workers.Start() // Inicia os workers
}

// Stop App
func (a *AppStruct) Stop(saveLastKey bool) {
	a.Keys.Stop()
	a.Workers.Stop()
	for {
		if !a.Workers.IsStarted {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	a.Ticker.Stop()
	a.Results.Stop()

	// Save Last Key
	if saveLastKey {
		a.LastKey.SetSaveLastKey(a.Carteira, fmt.Sprintf("%064x", a.Keys.GetLastKey()))
	}
	a.CtxCancel()
}
