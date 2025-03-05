package config

import (
	"time"

	"github.com/gagliardetto/solana-go"
)

const PRICE_UPDATE_INTERVAL = 1 * time.Second    // 1s
const CHART_UPDATE_INTERVAL = 60 * time.Second   // 60 = 1mins
const DEPOSIT_UPDATE_INTERVAL = 10 * time.Second // 10s

const ROUND_CRON_SPEC = "0 0 0 * * *"        // 매일 00시 00분 00초에 라운드 시작
const LEADER_BOARD_CRON_SPEC = "0 0 * * * *" // 매 시간 00분 00초에 리더보드 갱신

const SOLANA_RPC_URL = ""

var (
	DEPOSIT_ADDRESS, _ = solana.PublicKeyFromBase58("CoFf4ZpbTJRoPxdJ64JvMn4pVR1wjhvARc8ed91i9i37")
)

const USER_USD_AMOUNT = 10000
