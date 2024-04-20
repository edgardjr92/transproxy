package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/edgardjr92/transproxy/cache"
	"github.com/edgardjr92/transproxy/proxy"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

const (
	NetworkURL             = "NETWORK_URL"
	MonitorCheckIntervalMs = "MONITOR_CHECK_INTERVAL_MS"
	MaxWaitTimeMs          = "MAX_WAIT_TIME_MS"
	RpcPort                = "RPC_PORT"
)

type Config struct {
	NetworkURL           string
	MonitorCheckInterval time.Duration
	MaxWaitTime          time.Duration
	RpcPort              string
}

func main() {
	ctx := context.Background()

	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	// connect to the Ethereum client
	client, err := ethclient.Dial(config.NetworkURL)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	defer client.Close()

	cache := cache.New[string, types.Transaction]()

	monitor := proxy.NewMonitor(client, config.MonitorCheckInterval)
	proxysvc := proxy.NewProxy(client, cache, config.MaxWaitTime)

	monitor.Start(ctx, proxysvc.OnGasPriceChange)

	server := rpc.NewServer()
	if err := server.RegisterName("eth", proxysvc); err != nil {
		log.Fatalf("Failed to register RPC service: %v", err)
	}

	http.HandleFunc("/rpc", server.ServeHTTP)

	slog.Info("Starting server", "port", config.RpcPort)

	if err := http.ListenAndServe(":"+config.RpcPort, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func loadConfig() (Config, error) {
	err := godotenv.Load("./.env")

	if err != nil {
		return Config{}, err
	}

	checkInterval := fmt.Sprintf("%sms", getenv(MonitorCheckIntervalMs, "10000"))
	maxWaitTime := fmt.Sprintf("%sms", getenv(MaxWaitTimeMs, "300000"))

	return Config{
		NetworkURL:           getenv(NetworkURL, "http://localhost:7545"),
		MonitorCheckInterval: strToDuration(checkInterval),
		MaxWaitTime:          strToDuration(maxWaitTime),
		RpcPort:              getenv(RpcPort, "8081"),
	}, nil
}

func getenv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func strToDuration(str string) time.Duration {
	d, err := time.ParseDuration(str)
	if err != nil {
		log.Fatalf("Failed to parse duration: %v", err)
	}
	return d
}
