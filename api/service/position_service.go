package service

import (
	"errors"
	"tradedotdotfun-backend/api/db"
	"tradedotdotfun-backend/api/types"
	"tradedotdotfun-backend/common/model"
)

func CreatePosition(address string, req *types.CreatePositionRequest) (*types.CreatePositionResponse, error) {
	round := GetRound().Round
	price := GetPrice()

	// Calculate position size
	// Position size = Amount * Leverage / Price
	positionSize := req.Amount * float64(req.Leverage) / price[req.Token]
	// Calculate liquidation price
	// For long positions: entry_price * (1 - 1/leverage)
	// For short positions: entry_price * (1 + 1/leverage)
	var liquidationPrice float64
	if req.Type == "long" {
		liquidationPrice = price[req.Token] * (1 - 1/float64(req.Leverage))
	} else {
		liquidationPrice = price[req.Token] * (1 + 1/float64(req.Leverage))
	}

	db := db.GetConnection()
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	account := &model.Account{}
	if err := tx.Where("round = ? and deleted = false and address = ?", round, address).First(account).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if account.USDAmount < req.Amount {
		tx.Rollback()
		return nil, errors.New("insufficient balance")
	}

	account.USDAmount -= req.Amount

	position := &model.Position{
		Round:            uint64(round),
		Address:          address,
		Type:             string(req.Type),
		Leverage:         req.Leverage,
		Amount:           req.Amount,
		Token:            req.Token,
		EntryPrice:       price[req.Token],
		PositionSize:     positionSize,
		LiquidationPrice: liquidationPrice,
		Status:           model.StatusActive,
	}

	if err := tx.Create(position).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Save(account).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &types.CreatePositionResponse{Status: "success"}, nil
}

func ClosePosition(address string, positionId uint64, percentage float64) (*types.ClosePositionResponse, error) {
	round := GetRound().Round
	price := GetPrice()

	db := db.GetConnection()
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	position := &model.Position{}
	if err := tx.Where("id = ? AND status = ?", positionId, model.StatusActive).First(position).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if position.Address != address {
		tx.Rollback()
		return nil, errors.New("not position owner")
	}

	if position.Round != uint64(round) {
		tx.Rollback()
		return nil, errors.New("cannot close position in same round")
	}

	// Calculate PNL
	currentPrice := price[position.Token]
	currentValue := currentPrice * position.PositionSize
	initialAmount := position.Amount * float64(position.Leverage)
	var pnl float64
	if position.Type == "long" {
		pnl = currentValue - initialAmount
	} else {
		pnl = initialAmount - currentValue
	}
	if position.Amount+pnl > 0 {
		// 원금에 pnl을 더했을때도 음수인 경우는 청산 처리를 한다
		// Get account to update balance
		account := &model.Account{}
		if err := tx.Where("address = ?", address).First(account).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		// Update account balance
		account.USDAmount += (position.Amount + pnl) * percentage / 100

		if err := tx.Save(account).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Mark position as closed
	if percentage == 100 {
		position.Status = model.StatusClosed
	} else {
		position.Amount -= position.Amount * percentage / 100
		position.PositionSize -= position.PositionSize * percentage / 100
	}

	if err := tx.Save(position).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &types.ClosePositionResponse{Status: "success"}, nil
}

func GetPosition(round uint64, address string) types.GetPositionResponse {
	db := db.GetConnection()

	var positions []model.Position
	if err := db.Where("round = ? AND address = ? AND status = ?", round, address, model.StatusActive).Find(&positions).Error; err != nil {
		return nil
	}

	return ConvertPositionModelToDto(positions)
}

func ConvertPositionModelToDto(positions []model.Position) []types.Position {
	price := GetPrice()
	var result []types.Position
	for _, position := range positions {
		currentPrice := price[position.Token]
		currentValue := currentPrice * position.PositionSize
		initialAmount := position.Amount * float64(position.Leverage)
		var pnl float64
		if position.Type == "long" {
			pnl = currentValue - initialAmount
		} else {
			pnl = initialAmount - currentValue
		}
		roi := max(pnl/position.Amount, -1)

		result = append(result, types.Position{
			ID:               position.ID,
			Type:             position.Type,
			Leverage:         position.Leverage,
			Amount:           position.Amount,
			Token:            position.Token,
			EntryPrice:       position.EntryPrice,
			PositionSize:     position.PositionSize,
			LiquidationPrice: position.LiquidationPrice,
			PNL:              pnl,
			ROI:              roi,
			Created_dt:       position.Created_dt.Format("2006-01-02 15:04:05"),
		})
	}
	return result
}
