package proxy

import (
	"context"
	"log/slog"
	"math/big"
	"time"
)

type CallbackFn func(ctx context.Context, gasPrice *big.Int)

type Monitor struct {
	ethClient     EthClient
	checkInterval time.Duration
}

func NewMonitor(ethClient EthClient, checkInterval time.Duration) *Monitor {
	return &Monitor{
		ethClient:     ethClient,
		checkInterval: checkInterval,
	}
}

// Start starts the monitor to check the gas price at regular intervals.
func (m *Monitor) Start(ctx context.Context, cb CallbackFn) {
	go func() {
		for range time.Tick(m.checkInterval) {
			gasPrice, err := m.ethClient.SuggestGasPrice(ctx)

			if err != nil {
				slog.Error("faleid when calling eth_gasPrice", "err", err)
				continue
			}

			slog.Info("current gas price", "gasPrice", gasPrice.String())

			cb(ctx, gasPrice)
		}
	}()
}
