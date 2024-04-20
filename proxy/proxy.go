package proxy

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

type Proxy struct {
	ethClient   EthClient
	txCache     Cache
	maxWaitTime time.Duration
}

func NewProxy(ethClient EthClient, txCache Cache, maxWaitTime time.Duration) *Proxy {
	return &Proxy{
		ethClient:   ethClient,
		txCache:     txCache,
		maxWaitTime: maxWaitTime,
	}
}

// SendRawTransaction sends a raw transaction to the connected Ethereum node. (eth_sendRawTransaction)
func (p *Proxy) SendRawTransaction(ctx context.Context, txData string) (common.Hash, error) {
	slog.Info("SendRawTransaction", "txData", txData)

	tx, err := RawTransaction(txData).Decode()
	if err != nil {
		slog.Error("failed to decode transaction", "err", err)
		return common.Hash{}, err
	}

	senderAcc, err := p.readSenderAccount(&tx)
	if err != nil {
		slog.Error("failed to read sender transaction", "err", err)
		return common.Hash{}, err
	}

	if err := p.scheduleTransaction(tx, senderAcc); err != nil {
		slog.Error("failed to schedule transaction", "err", err)
		return common.Hash{}, err
	}

	return tx.Hash(), nil
}

// ScheduleTransaction schedules a transaction to be sent to the connected Ethereum node.
// The transaction will be sent when the gas price is lower than the transaction's gas price.
// If another transaction with the same sender and nonce is scheduled, the previous transaction is canceled.
func (p *Proxy) scheduleTransaction(tx types.Transaction, senderAcc string) error {
	scheduleID := fmt.Sprintf("%s|%d", senderAcc, tx.Nonce())

	if _, ok := p.txCache.Get(scheduleID); ok {
		p.txCache.Remove(scheduleID)
		slog.Info("transaction canceled", "hash", tx.Hash())
		return nil
	}

	p.txCache.Set(scheduleID, tx, p.maxWaitTime)
	slog.Info("transaction scheduled", "hash", tx.Hash())

	return nil
}

// OnGasPriceChange is called when the gas price changes.
// If the gas price is lower than the transaction's gas price, the transaction is sent to the connected Ethereum node.
func (p *Proxy) OnGasPriceChange(ctx context.Context, gasPrice *big.Int) {
	for key, tx := range p.txCache.Values() {
		if tx.GasPrice().Cmp(gasPrice) >= 0 {
			if err := p.ethClient.SendTransaction(ctx, &tx); err != nil {
				slog.Error("failed when calling eth_sendRawTransaction", "err", err)
			}
			slog.Info("transaction sent", "hash", tx.Hash())
			p.txCache.Remove(key)
		}
	}
}

// readSenderAccount reads the sender account from the given transaction.
func (Proxy) readSenderAccount(tx *types.Transaction) (string, error) {
	signer := types.NewEIP155Signer(tx.ChainId())
	sender, err := signer.Sender(tx)

	if err != nil {
		return "", err
	}
	return sender.Hex(), err
}

// ChainId returns the chain ID of the connected Ethereum node.
// It is necessary to integrate with Metamask.
func (p *Proxy) ChainId(ctx context.Context) (hexutil.Big, error) {
	slog.Info("ChainId")

	chainID, err := p.ethClient.ChainID(ctx)
	if err != nil {
		slog.Error("failed when calling eth_chainId", "err", err)
		return hexutil.Big{}, err
	}

	return hexutil.Big(*chainID), nil
}

// BlockNumber returns the latest block number of the connected Ethereum node.
// It is necessary to integrate with Metamask.
func (p *Proxy) BlockNumber(ctx context.Context) (hexutil.Uint64, error) {
	slog.Info("BlockNumber")

	blockNumber, err := p.ethClient.BlockNumber(ctx)
	if err != nil {
		slog.Error("failed when calling eth_blockNumber", "err", err)
		return hexutil.Uint64(0), err
	}

	return hexutil.Uint64(blockNumber), nil
}

// GetBlockByNumber returns the block with the given number from the connected Ethereum node.
// It is necessary to integrate with Metamask.
func (p *Proxy) GetBlockByNumber(ctx context.Context, number hexutil.Big, fullTx bool) (*types.Block, error) {
	slog.Info("GetBlockByNumber", "blockNumber", number, "fullTx", fullTx)

	block, err := p.ethClient.BlockByNumber(ctx, number.ToInt())
	if err != nil {
		slog.Error("failed when calling eth_getBlockByNumber", "err", err)
		return nil, err
	}

	return block, nil
}

// GetBalance returns the balance of the given address from the connected Ethereum node.
// It is necessary to integrate with Metamask.
func (p *Proxy) GetBalance(ctx context.Context, address common.Address, blockNumber *hexutil.Big) (hexutil.Big, error) {
	slog.Info("GetBalance", "address", address, "blockNumber", blockNumber)

	balance, err := p.ethClient.BalanceAt(ctx, address, blockNumber.ToInt())
	if err != nil {
		slog.Error("failed when calling eth_getBalance", "err", err)
		return hexutil.Big{}, err
	}
	return hexutil.Big(*balance), nil
}

// GetTransactionCount returns the number of transactions sent from the given address from the connected Ethereum node.
// It is necessary to integrate with Metamask.
func (p *Proxy) GetTransactionCount(ctx context.Context, address common.Address, blockNumber string) (*hexutil.Uint64, error) {
	slog.Info("GetTransactionCount", "address", address, "blockNumber", blockNumber)
	var count *hexutil.Uint64

	if err := p.ethClient.Client().CallContext(ctx, &count, "eth_getTransactionCount", address, blockNumber); err != nil {
		slog.Error("failed when calling eth_getTransactionCount", "err", err)
		return count, err
	}

	return count, nil
}

// GasPrice returns the current gas price of the connected Ethereum node.
// It is necessary to integrate with Metamask.
func (p *Proxy) GasPrice(ctx context.Context) (hexutil.Big, error) {
	slog.Info("GasPrice")

	gasPrice, err := p.ethClient.SuggestGasPrice(ctx)
	if err != nil {
		slog.Error("failed when calling eth_gasPrice", "err", err)
		return hexutil.Big{}, err
	}

	return hexutil.Big(*gasPrice), nil
}
