//go:build !DEV
// +build !DEV

package config

import "time"

const RPC_LATEST_CACHE_INTERVAL = 1 * time.Second // 1s

const SQLITE_DB_PATH = "/home/ubuntu/tradedotdotfun.db"

const INITIAL_USD_BALANCE = float64(10000)

var COIN_LIST = []string{
	"BTCUSDT",
	"ETHUSDT",
	"SOLUSDT",
}
