package application

import (
	"btcgo/internal/core"
	"btcgo/internal/domain"
	"btcgo/internal/utils"
	"fmt"
	"math/big"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
)

type Application struct {
	wallets        *domain.Wallets
	cpusNumber     int
	privateKeyChan chan *big.Int
	resultChan     chan *big.Int
	modSelected    int
	savedWallet    string
	privateKeyInt  *big.Int
}

func New(wallets *domain.Wallets, cpusNumber int, modSelected int, savedWallet string, privateKeyInt *big.Int) *Application {
	return &Application{
		wallets:        wallets,
		cpusNumber:     cpusNumber,
		privateKeyChan: make(chan *big.Int),
		resultChan:     make(chan *big.Int),
		modSelected:    modSelected,
		savedWallet:    savedWallet,
		privateKeyInt:  privateKeyInt,
	}
}

func (a *Application) Run() {
	defer a.closeChannels()

	var (
		wg                = &sync.WaitGroup{}
		startTime         = time.Now()
		keysChecked int64 = 0
	)

	// Start worker goroutines
	for i := 0; i < a.cpusNumber; i++ {
		wg.Add(1)
		go core.Worker(a.wallets, a.privateKeyChan, a.resultChan, wg)
	}

	go func() {
		for {
			a.privateKeyChan <- a.privateKeyInt
			a.privateKeyInt.Add(a.privateKeyInt, big.NewInt(1))
			atomic.AddInt64(&keysChecked, 1)
		}
	}()

	a.logWithOption(startTime, &keysChecked)

	a.verifyResultChan()
	wg.Wait()

	elapsedTime := time.Since(startTime).Seconds()
	keysPerSecond := float64(keysChecked) / elapsedTime
	fmt.Printf("Chaves checadas: %s\n", humanize.Comma(int64(keysChecked)))
	fmt.Printf("Tempo: %.2f seconds\n", elapsedTime)
	fmt.Printf("Chaves por segundo: %s\n", humanize.Comma(int64(keysPerSecond)))
}

func (a *Application) verifyResultChan() {
	select {
	case foundAddress := <-a.resultChan:
		wif := utils.GenerateWif(foundAddress)
		color.Yellow("Chave privada encontrada: %064x\n", foundAddress)
		color.Yellow("WIF: %s", wif)

		if a.modSelected == 2 {
			foundAddressString := fmt.Sprintf("%064x", foundAddress)
			_ = utils.SaveLastKeyWallet("ultimaChavePorCarteira.txt", a.savedWallet, foundAddressString)
		}

		file, err := os.OpenFile("chaves_encontradas.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		currentTime := time.Now().Format("2006-01-02 15:04:05")

		if err != nil {
			color.Red("Erro ao abrir o arquivo chaves_encontradas.txt")
			_, err := file.WriteString(fmt.Sprintf("Data/Hora: %s | Chave privada: %064x | WIF: %s\n", currentTime, foundAddress, wif))
			if err != nil {
				color.Red("Erro ao escrever no arquivo chaves_encontradas.txt")
				return
			}

		}
		file.WriteString(fmt.Sprintf("Data/Hora: %s | Chave privada: %064x | WIF: %s\n", currentTime, foundAddress, wif))
		defer file.Close()
	}

}

func (a *Application) logWithOption(startTime time.Time, keysChecked *int64) {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		tickerPersist := time.NewTicker(5 * time.Second)

		defer ticker.Stop()
		defer tickerPersist.Stop()

		for {
			select {
			case <-ticker.C:
				elapsedTime := time.Since(startTime).Seconds()
				count := atomic.LoadInt64(keysChecked)

				keysPerSecond := float64(count) / elapsedTime
				fmt.Printf("Chaves checadas: %s Chaves por segundo: %s\n", humanize.Comma(int64(count)), humanize.Comma(int64(keysPerSecond)))

			case <-tickerPersist.C:
				lastKey := fmt.Sprintf("%064x", a.privateKeyInt)
				err := utils.SaveLastKeyWallet("ultimaChavePorCarteira.txt", a.savedWallet, lastKey)
				if err != nil {
					color.Red("Erro ao salvar a ultima chave, err: %v\n", err)
				}
			}
		}
	}()
}

func (a *Application) closeChannels() {
	close(a.privateKeyChan)
	close(a.resultChan)
}
