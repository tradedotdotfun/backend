package service

import (
	"log"
	"strconv"
	"time"
	"tradedotdotfun-backend/api/config"
	"tradedotdotfun-backend/api/db"
	"tradedotdotfun-backend/api/types"
	commonconfig "tradedotdotfun-backend/common/config"
	"tradedotdotfun-backend/common/model"
)

var priceMap map[string]float64

func GetPrice() types.GetPriceResponse {
	return priceMap
}

func StartPriceUpdater() {
	prices, err := FetchPriceFromDB(commonconfig.COIN_LIST)
	if err != nil {
		log.Println(err)
		return
	}
	priceMap = prices
	ticker := time.NewTicker(config.PRICE_REFRESH_INTERVAL)
	go func() {
		for {
			<-ticker.C
			prices, err := FetchPriceFromDB(commonconfig.COIN_LIST)
			if err != nil {
				log.Println(err)
				continue
			}
			priceMap = prices
		}
	}()
}

func FetchPriceFromDB(coinIDs []string) (map[string]float64, error) {
	conn := db.GetConnection()

	var keyValueStores []model.KeyValueStore
	err := conn.Where("key IN ?", coinIDs).Find(&keyValueStores).Error
	if err != nil {
		return nil, err
	}

	newPriceMap := make(map[string]float64)
	for _, keyValueStore := range keyValueStores {
		price, err := strconv.ParseFloat(keyValueStore.Value, 64)
		if err != nil {
			return nil, err
		}
		newPriceMap[keyValueStore.Key] = price
	}
	return newPriceMap, nil
}
