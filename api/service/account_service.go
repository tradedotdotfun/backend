package service

import (
	"errors"
	"tradedotdotfun-backend/api/db"
	"tradedotdotfun-backend/api/types"
	"tradedotdotfun-backend/common/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetAccount(round uint64, address string) *types.GetAccountResponse {
	db := db.GetConnection()

	var name string
	err := db.Model(&model.AddressName{}).
		Select("name").
		Where("address = ?", address).
		First(&name).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}

	var account model.Account
	var positions []model.Position

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("round = ? AND address = ? AND status = ?", round, address, model.StatusActive).Find(&positions).Error; err != nil {
			return err
		}

		if err := tx.Where("round = ? AND address = ? AND deleted = false", round, address).First(&account).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil
	}

	var leaderBoard model.LeaderBoard
	err = db.Model(&model.LeaderBoard{}).
		Where("round = ? AND address = ?", round, address).
		First(&leaderBoard).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}

	positionDto := ConvertPositionModelToDto(positions)
	var totalAmount float64
	for _, position := range positionDto {
		totalAmount += position.PNL + position.Amount
	}

	return &types.GetAccountResponse{
		Name:                 name,
		Rank:                 leaderBoard.Rank,
		USDAmount:            account.USDAmount,
		EstimatedTotalAmount: account.USDAmount + totalAmount,
	}
}

func AddName(address string, name string) (*types.AddNameResponse, error) {
	db := db.GetConnection()

	addressName := model.AddressName{
		Address: address,
		Name:    name,
	}

	if err := db.Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{"name"}),
	}).Create(&addressName).Error; err != nil {
		return nil, err
	}

	return &types.AddNameResponse{Status: "success"}, nil
}
