package proxy

import (
	"context"
	big "math/big"
	"testing"
	"time"

	"github.com/edgardjr92/transproxy/cache"
	"github.com/ethereum/go-ethereum/core/types"
	mock "github.com/stretchr/testify/mock"
)

const (
	txData = "0xf869018203e882520894f17f52151ebef6c7334fad080c5704d77216b732881bc16d674ec80000801ba02da1c48b670996dcb1f447ef9ef00b33033c48a4fe938f420bec3e56bfd24071a062e0aa78a81bf0290afbc3a9d8e9a068e6d74caa66c5e0fa8a46deaae96b0833"
	txHash = "0xb40ca2b75e7312bfa842991b14bc7cac495736a4fcf18e70240c9702525c2ade"
)

func TestSendRawTransaction(t *testing.T) {
	ctx := context.TODO()
	cache := cache.New[string, types.Transaction]()
	ethClientMock := NewEthClientMock(t)
	proxy := NewProxy(ethClientMock, cache, 30*time.Minute)
	t.Run("should schedule transaction when it is not scheduled", func(t *testing.T) {
		hash, err := proxy.SendRawTransaction(ctx, txData)
		defer cache.Clear()

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if hash.String() != txHash {
			t.Errorf("Expected %s, got %s", txHash, hash.String())
		}

		if _, ok := cache.Get("0x627306090abaB3A6e1400e9345bC60c78a8BEf57|1"); !ok {
			t.Errorf("Expected transaction to be scheduled")
		}
	})

	t.Run("should cancel transaction when it is already scheduled", func(t *testing.T) {
		// first transaction
		hash, err := proxy.SendRawTransaction(ctx, txData)
		defer cache.Clear()

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if hash.String() != txHash {
			t.Errorf("Expected %s, got %s", txHash, hash.String())
		}

		if _, ok := cache.Get("0x627306090abaB3A6e1400e9345bC60c78a8BEf57|1"); !ok {
			t.Errorf("Expected transaction to be scheduled")
		}

		// second transaction
		hash, err = proxy.SendRawTransaction(ctx, txData)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if hash.String() != txHash {
			t.Errorf("Expected %s, got %s", txHash, hash.String())
		}

		if _, ok := cache.Get("0x627306090abaB3A6e1400e9345bC60c78a8BEf57|1"); ok {
			t.Errorf("Expected transaction to be canceled")
		}
	})
}

func TestOnGasPriceChange(t *testing.T) {
	ctx := context.TODO()
	cache := cache.New[string, types.Transaction]()
	ethClientMock := NewEthClientMock(t)
	proxy := NewProxy(ethClientMock, cache, 30*time.Minute)

	t.Run("should send transaction when gas price is lower than transaction's gas price", func(t *testing.T) {
		defer cache.Clear()

		currentGasPrice := big.NewInt(1000000000)
		txGasPrice := big.NewInt(1000000001)
		tx := types.NewTx(&types.LegacyTx{GasPrice: txGasPrice})

		cache.Set("0x627306090abaB3A6e1400e9345bC60c78a8BEf57|1", *tx, 30*time.Minute)

		ethClientMock.On(
			"SendTransaction",
			mock.AnythingOfType("context.todoCtx"),
			tx,
		).Return(nil).Times(1)

		proxy.OnGasPriceChange(ctx, currentGasPrice)

		ethClientMock.AssertExpectations(t)

		if _, ok := cache.Get("0x627306090abaB3A6e1400e9345bC60c78a8BEf57|1"); ok {
			t.Errorf("Expected transaction to be removed from cache")
		}
	})

	t.Run("should send transaction when gas price is equal to transaction's gas price", func(t *testing.T) {
		defer cache.Clear()

		currentGasPrice := big.NewInt(1000000000)
		txGasPrice := big.NewInt(1000000000)
		tx := types.NewTx(&types.LegacyTx{GasPrice: txGasPrice})

		cache.Set("0x627306090abaB3A6e1400e9345bC60c78a8BEf57|1", *tx, 30*time.Minute)

		ethClientMock.On(
			"SendTransaction",
			mock.AnythingOfType("context.todoCtx"),
			tx,
		).Return(nil).Times(1)

		proxy.OnGasPriceChange(ctx, currentGasPrice)

		ethClientMock.AssertExpectations(t)

		if _, ok := cache.Get("0x627306090abaB3A6e1400e9345bC60c78a8BEf57|1"); ok {
			t.Errorf("Expected transaction to be removed from cache")
		}
	})

	t.Run("should not send transaction when gas price is higher than transaction's gas price", func(t *testing.T) {
		defer cache.Clear()

		currentGasPrice := big.NewInt(1000000001)
		txGasPrice := big.NewInt(1000000000)
		tx := types.NewTx(&types.LegacyTx{GasPrice: txGasPrice})

		cache.Set("0x627306090abaB3A6e1400e9345bC60c78a8BEf57|1", *tx, 30*time.Minute)

		proxy.OnGasPriceChange(ctx, currentGasPrice)

		ethClientMock.AssertNotCalled(t, "SendTransaction", mock.AnythingOfType("context.todoCtx"), tx)

		if _, ok := cache.Get("0x627306090abaB3A6e1400e9345bC60c78a8BEf57|1"); !ok {
			t.Errorf("Expected transaction to be in cache")
		}
	})
}
