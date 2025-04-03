package cache

import (
	"log"
	"strconv"
	"tradedotdotfun-backend/common/config"
	"tradedotdotfun-backend/common/db"
	"tradedotdotfun-backend/common/model"
)

var round uint64 = 0
var roundStatus model.RoundStatus = model.RoundStatusReady
var priceMap = make(map[string]float64)

func Init() {
	round = fetchRoundFromDB()
	roundStatus = fetchRoundStatusFromDB()
	priceMap = FetchPriceFromDB(config.COIN_LIST)
}

func fetchRoundFromDB() uint64 {
	conn := db.GetConnection()

	var currentRound uint64
	if err := conn.Model(&model.KeyValueStore{}).
		Select("value").
		Where("key = ?", model.CURRENT_ROUND_KEY).
		Find(&currentRound).Error; err != nil {
		log.Println(err)
		return 0
	}
	return currentRound
}

func fetchRoundStatusFromDB() model.RoundStatus {
	conn := db.GetConnection()

	var currentRoundStatus model.RoundStatus
	if err := conn.Model(&model.KeyValueStore{}).
		Select("value").
		Where("key = ?", model.CURRENT_ROUND_STATUS_KEY).
		Find(&currentRoundStatus).Error; err != nil {
		log.Println(err)
		return model.RoundStatusReady
	}
	return currentRoundStatus
}

func FetchPriceFromDB(coinIDs []string) map[string]float64 {
	conn := db.GetConnection()

	var keyValueStores []model.KeyValueStore
	err := conn.Where("key IN ?", coinIDs).Find(&keyValueStores).Error
	if err != nil {
		return nil
	}

	newPriceMap := make(map[string]float64)
	for _, keyValueStore := range keyValueStores {
		price, err := strconv.ParseFloat(keyValueStore.Value, 64)
		if err != nil {
			return nil
		}
		newPriceMap[keyValueStore.Key] = price
	}
	return newPriceMap
}

func GetPrice() map[string]float64 {
	return priceMap
}

func SetPrice(prices map[string]float64) {
	priceMap = prices
}

func GetRound() uint64 {
	return round
}

func SetRound(newRound uint64) {
	round = newRound
}

func GetRoundStatus() model.RoundStatus {
	return roundStatus
}

func SetRoundStatus(newRoundStatus model.RoundStatus) {
	roundStatus = newRoundStatus
}
