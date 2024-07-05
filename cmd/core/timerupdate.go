/**
 * BTCGO
 *
 * Modulo : Time Ticker Update
 */

package core

import (
	"context"
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
)

type TimerUpdater struct {
	Ctx       context.Context
	Ticker    *time.Ticker
	StartTime time.Time
	IsStarted bool
}

func NewTicker(ctx context.Context) *TimerUpdater {
	return &TimerUpdater{
		Ctx:       ctx,
		IsStarted: false,
	}
}

// Start
func (t *TimerUpdater) Start(timerSeconds int) {
	t.IsStarted = true
	t.StartTime = time.Now()
	t.Ticker = time.NewTicker(time.Duration(timerSeconds) * time.Second)
	t.tickerProcessor()
}

// Stop
func (t *TimerUpdater) Stop() {
	t.Ticker.Stop()
	t.IsStarted = false
}

// Ticker
func (t TimerUpdater) tickerProcessor() {
	go func() {
		for {
			select {
			case <-t.Ticker.C:
				keyscheck := App.Keys.GetTotalKeys()
				elapsedTime := time.Since(t.StartTime).Seconds()
				fmt.Printf("Chaves checadas: %s Chaves por segundo: %s\n", humanize.Comma(int64(keyscheck)), humanize.Comma(int64(keyscheck/elapsedTime)))
				App.LastKey.SetSaveLastKey(App.Carteira, fmt.Sprintf("%064x", App.Keys.GetLastKey()))

			case <-t.Ctx.Done():
				t.Stop()
				return
			}
		}
	}()
}
