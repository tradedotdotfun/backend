package config

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/joho/godotenv"
)

const RPC_LATEST_CACHE_INTERVAL = 1 * time.Second // 1s

const INITIAL_USD_BALANCE = float64(10000)

var COIN_LIST = []string{
	"BTCUSDT",
	"ETHUSDT",
	"SOLUSDT",
	"SONICUSDT",
}

var (
	RPC_URL         string
	DEPOSIT_ADDRESS solana.PublicKey
	SQLITE_DB_PATH  string
	ADMIN_PUBKEY    string
)

func init() {
	// 테스트 환경이면 init 실행 안 함
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") {
			log.Println("Skipping init() in test mode")
			return
		}
	}

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	adminPubkey := os.Getenv("ADMIN_PUBKEY")
	if adminPubkey == "" {
		log.Fatal("ADMIN_PUBKEY is not set in the environment variables")
	}
	ADMIN_PUBKEY = adminPubkey

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		log.Fatal("DB_PATH is not set in the environment variables")
	}
	SQLITE_DB_PATH = dbPath

	rpcUrl := os.Getenv("RPC_URL")
	if rpcUrl == "" {
		log.Fatal("RPC_URL is not set in the environment variables")
	}
	RPC_URL = rpcUrl

	depositAddress := os.Getenv("DEPOSIT_ADDRESS")
	if depositAddress == "" {
		log.Fatal("DEPOSIT_ADDRESS is not set in the environment variables")
	}
	depositAddressStr, err := solana.PublicKeyFromBase58(depositAddress)
	if err != nil {
		log.Fatalf("Failed to parse deposit address: %v", err)
	}
	DEPOSIT_ADDRESS = depositAddressStr
}
