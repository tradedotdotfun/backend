package model

const CURRENT_ROUND_KEY = "current_round"
const CURRENT_ROUND_STATUS_KEY = "current_round_status"
const LAST_TX_SIGNATURE_KEY = "last_tx_signature"

type RoundStatus string

const (
	RoundStatusStart RoundStatus = "start"
	RoundStatusEnd   RoundStatus = "end"
	RoundStatusReady RoundStatus = "ready"
)
