package proxy

import (
	context "context"
	big "math/big"
	"testing"
	"time"

	mock "github.com/stretchr/testify/mock"
)

func TestMonitor_Start(t *testing.T) {
	// Create a new Monitor instance
	ctx := context.TODO()
	ethClientMock := NewEthClientMock(t)
	monitor := &Monitor{
		checkInterval: time.Second,
		ethClient:     ethClientMock,
	}

	ethClientMock.On("SuggestGasPrice", mock.AnythingOfType("context.todoCtx")).
		Return(big.NewInt(100), nil).Times(1)

	resultCh := make(chan *big.Int)

	expectedGasPrice := big.NewInt(100)

	callbackFn := func(ctx context.Context, gasPrice *big.Int) {
		resultCh <- gasPrice
	}

	go monitor.Start(ctx, callbackFn)

	select {
	case gasPrice := <-resultCh:
		if gasPrice.Cmp(expectedGasPrice) != 0 {
			t.Errorf("Unexpected gas price. Expected: %s, Got: %s", expectedGasPrice.String(), gasPrice.String())
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for callback result")
	}
}
