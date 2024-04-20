package proxy

import (
	"context"
	"log/slog"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

type EthClient interface {
	SendTransaction(ctx context.Context, tx *types.Transaction) error
	ChainID(ctx context.Context) (*big.Int, error)
	BlockNumber(ctx context.Context) (uint64, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	BalanceAt(ctx context.Context, address common.Address, blockNumber *big.Int) (*big.Int, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	Client() *rpc.Client
}

type Cache interface {
	Get(key string) (types.Transaction, bool)
	Set(key string, value types.Transaction, ttl time.Duration)
	Remove(key string)
	Values() map[string]types.Transaction
}

type RawTransaction string

func (r RawTransaction) Decode() (types.Transaction, error) {
	txBytes := common.FromHex(string(r))
	tx := new(types.Transaction)

	if err := tx.UnmarshalBinary(txBytes); err != nil {
		slog.Error("Failed to decode transaction", "err", err)
		return types.Transaction{}, err
	}

	return *tx, nil
}
