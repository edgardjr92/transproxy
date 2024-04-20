# Transproxy JSON RPC Server

This Ethereum Proxy JSON RPC Server acts as a middleware between clients and the Ethereum network. It provides a unique feature where transactions are scheduled and only submitted when the network's gas price is equal to or less than the specified transaction price. This server ensures that transactions with the same nonce will cancel any previously scheduled transactions, preventing them from being submitted to the mempool if a new transaction is sent.

## Environment Variables

- `NETWORK_URL`: URL of the Ethereum network the server will proxy to.
- `MONITOR_CHECK_INTERVAL_MS`: Interval in milliseconds for checking the network gas price.
- `MAX_WAIT_TIME_MS`: Maximum wait time in milliseconds for a transaction to be scheduled before dropping if the desired gas price is not met.
- `RPC_PORT`: Port for the RPC server.

## Getting Started

### Prerequisites

- Docker
- Docker Compose

### Running the App via Docker Compose

1. Create a `docker-compose.yml` file in your project root

```sh
docker-compose up -d
```
### Example Usage

Send a transaction to the server using curl:

```sh
curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_sendRawTransaction","params":["0xf86b808504a817c80082520894a0b86991c6218b36c1d19d4a2e9eb0ce3606eb48b"],"id":1}' http://localhost:8080/rpc
```
## Running Tests

```sh
go test ./...
```
