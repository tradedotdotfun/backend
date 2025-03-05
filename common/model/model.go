package model

import (
	"strconv"
	"time"
)

type KeyValueStore struct {
	ID         uint64    `gorm:"primary_key"`
	Key        string    `gorm:"column:key;uniqueIndex:idx1_key_value_stores"`
	Value      string    `gorm:"column:value"`
	Created_dt time.Time `gorm:"column:created_dt;autoCreateTime"`
	Updated_dt time.Time `gorm:"column:updated_dt;autoUpdateTime"`
}

func (k *KeyValueStore) GetValueAsInt64OrDefault(_default int) int64 {
	result, err := strconv.ParseInt(k.Value, 10, 64)
	if err != nil {
		return int64(_default)
	}
	return result
}

type ChartData struct {
	ID        int64     `gorm:"primary_key"`
	Symbol    string    `gorm:"column:symbol;index:idx1_chart_data,priority:1"`
	OpenTime  time.Time `gorm:"column:open_time"`
	CloseTime time.Time `gorm:"column:close_time;index:idx1_chart_data,priority:2"`
	Open      float64   `gorm:"column:open"`
	High      float64   `gorm:"column:high"`
	Low       float64   `gorm:"column:low"`
	Close     float64   `gorm:"column:close"`
	Volume    float64   `gorm:"column:volume"`
}

type PositionStatus string

const (
	StatusActive     PositionStatus = "active"
	StatusClosed     PositionStatus = "closed"
	StatusLiquidated PositionStatus = "liquidated"
)

type Position struct {
	ID               uint64         `gorm:"primary_key"`
	Round            uint64         `gorm:"column:round;index:idx1_positions,priority:1;index:idx2_positions,priority:2"`
	Address          string         `gorm:"column:address;index:idx1_positions,priority:3"`
	Type             string         `gorm:"column:type;index:idx1_positions,priority:4;index:idx2_positions,priority:4"`
	Leverage         uint16         `gorm:"column:leverage"`
	Amount           float64        `gorm:"column:amount"`
	Token            string         `gorm:"column:token;index:idx2_positions,priority:3"`
	EntryPrice       float64        `gorm:"column:entry_price"`
	PositionSize     float64        `gorm:"column:position_size"`
	LiquidationPrice float64        `gorm:"column:liquidation_price;index:idx2_positions,priority:5"`
	Status           PositionStatus `gorm:"column:status;index:idx1_positions,priority:2;index:idx2_positions,priority:1"`
	Created_dt       time.Time      `gorm:"column:created_dt;autoCreateTime"`
	Updated_dt       time.Time      `gorm:"column:updated_dt;autoUpdateTime"`
}

type Account struct {
	ID         uint64    `gorm:"primary_key"`
	Round      uint64    `gorm:"column:round;index:idx1_addresses,priority:1"`
	Address    string    `gorm:"column:address;index:idx1_addresses,priority:3"`
	USDAmount  float64   `gorm:"column:usd_amount"`
	Deleted    bool      `gorm:"column:deleted;index:idx1_addresses,priority:2"`
	Created_dt time.Time `gorm:"column:created_dt;autoCreateTime"`
	Updated_dt time.Time `gorm:"column:updated_dt;autoUpdateTime"`
}

type AddressName struct {
	Address    string    `gorm:"column:address;primary_key"`
	Name       string    `gorm:"column:name"`
	Created_dt time.Time `gorm:"column:created_dt;autoCreateTime"`
	Updated_dt time.Time `gorm:"column:updated_dt;autoUpdateTime"`
}

type LeaderBoard struct {
	ID         uint64    `gorm:"primary_key"`
	Round      uint64    `gorm:"column:round;index:idx1_leader_boards,priority:1"`
	Address    string    `gorm:"column:address;index:idx2_leader_boards,priority:1"`
	Rank       uint64    `gorm:"column:rank;index:idx1_leader_boards,priority:2"`
	PnL        float64   `gorm:"column:pnl"`
	RoI        float64   `gorm:"column:roi"`
	Created_dt time.Time `gorm:"column:created_dt;autoCreateTime"`
	Updated_dt time.Time `gorm:"column:updated_dt;autoUpdateTime"`
}

type DepositEvent struct {
	Signature  string    `gorm:"column:signature;primary_key"`
	Slot       uint64    `gorm:"column:slot;index:idx1_deposit_events,priority:1"`
	Address    string    `gorm:"column:address"`
	Created_dt time.Time `gorm:"column:created_dt;autoCreateTime"`
	Updated_dt time.Time `gorm:"column:updated_dt;autoUpdateTime"`
}
