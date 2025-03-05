package liquidation

import (
	"fmt"
	"tradedotdotfun-backend/common/model"
	"tradedotdotfun-backend/indexer/cache"
	"tradedotdotfun-backend/indexer/db"

	"github.com/gofiber/fiber/v2/log"
)

func FindAndLiquidate() {
	round := cache.GetRound()
	prices := cache.GetPrice()

	for coinID, price := range prices {
		for _, positionType := range []string{"long", "short"} {
			go liquidate(round, coinID, positionType, price)
		}
	}
}

func liquidate(round uint64, coinID string, positionType string, price float64) {
	db := db.GetConnection()

	var liquidationQuery string
	if positionType == "long" {
		liquidationQuery = "liquidation_price >= ?"
	} else {
		liquidationQuery = "liquidation_price <= ?"
	}

	var positions []model.Position
	if err := db.Model(&model.Position{}).
		Where("round = ? AND token = ? AND type = ? AND status = ?", round, coinID, positionType, model.StatusActive).
		Where(liquidationQuery, price).
		Find(&positions).Error; err != nil {
		log.Errorf("Liquidation failed: %v", err)
		return
	}
	for _, position := range positions {
		fmt.Println(position)
	}

	if err := db.Model(&model.Position{}).
		Where("round = ? AND token = ? AND type = ? AND status = ?", round, coinID, positionType, model.StatusActive).
		Where(liquidationQuery, price).
		Update("status", model.StatusLiquidated).Error; err != nil {
		log.Errorf("Liquidation failed: %v", err)
		return
	}
}
