package types

import (
	"errors"
	"strings"
	"tradedotdotfun-backend/common/config"
)

type PositionType string

const (
	Long  PositionType = "long"
	Short PositionType = "short"
)

type CreatePositionRequest struct {
	Type     PositionType `json:"type" validate:"oneof=long short"`
	Leverage uint16       `json:"leverage"`
	Amount   float64      `json:"amount"`
	Token    string       `json:"token"`
}

func (c *CreatePositionRequest) Validate() error {
	if c.Leverage < 1 || c.Leverage > 100 {
		return errors.New("leverage must be between 1 and 100")
	}
	if c.Amount == 0 {
		return errors.New("amount must be greater than 0")
	}
	c.Token = strings.ToUpper(c.Token)
	for _, coin := range config.COIN_LIST {
		if coin == c.Token {
			return nil
		}
	}
	return errors.New("token not found in coin list")
}

type CreatePositionResponse struct {
	Status string `json:"status"`
}

type ClosePositionRequest struct {
	Percentage float64 `json:"percentage"`
}

func (c *ClosePositionRequest) Validate() error {
	if c.Percentage < 0 || c.Percentage > 100 {
		return errors.New("percentage must be between 0 and 100")
	}
	return nil
}

type ClosePositionResponse struct {
	Status string `json:"status"`
}

type GetPositionResponse []Position

type Position struct {
	ID               uint64  `json:"id"`
	Type             string  `json:"type"`
	Leverage         uint16  `json:"leverage"`
	Amount           float64 `json:"amount"`
	Token            string  `json:"token"`
	EntryPrice       float64 `json:"entry_price"`
	PositionSize     float64 `json:"position_size"`
	LiquidationPrice float64 `json:"liquidation_price"`
	PNL              float64 `json:"pnl"`
	ROI              float64 `json:"roi"`
	Created_dt       string  `json:"created_dt"`
}
