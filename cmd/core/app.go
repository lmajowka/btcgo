/**
 *	App BTCGO
 */

package core

import (
	"btcgo/cmd/utils"
	"context"
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

	// Request Prompts
	App.consolePrompts()

	// Set Number Max of CPUs
	App.setCPUs()

	// Start
	App.start()

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
	a.CtxCancel()
}
