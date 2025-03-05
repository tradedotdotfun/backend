package service

import (
	"log"
	"strings"
	"time"

	"tradedotdotfun-backend/api/config"
	"tradedotdotfun-backend/api/db"
	"tradedotdotfun-backend/api/types"
	commonconfig "tradedotdotfun-backend/common/config"
	"tradedotdotfun-backend/common/model"

	"github.com/patrickmn/go-cache"
)

var chartCache = cache.New(10*time.Minute, 10*time.Minute)

func GetChart(coinID string) types.GetChartResponse {
	coinID = strings.ToUpper(coinID)
	chartData, found := chartCache.Get(coinID)
	if found {
		return chartData.([]types.ChartData)
	}
	chartData, err := FetchChartDataFromDB(coinID)
	if err != nil {
		log.Println(err)
		return types.GetChartResponse{}
	}
	chartCache.Set(coinID, chartData, cache.DefaultExpiration)
	return chartData.([]types.ChartData)
}

func StartChartUpdater() {
	for _, coinID := range commonconfig.COIN_LIST {
		chartData, err := FetchChartDataFromDB(coinID)
		if err != nil {
			log.Println(err)
			return
		}
		chartCache.Set(coinID, chartData, cache.DefaultExpiration)
	}

	ticker := time.NewTicker(config.CHART_REFRESH_INTERVAL)
	go func() {
		for {
			<-ticker.C
			for _, coinID := range commonconfig.COIN_LIST {
				chartData, err := FetchChartDataFromDB(coinID)
				if err != nil {
					log.Println(err)
					continue
				}
				chartCache.Set(coinID, chartData, cache.DefaultExpiration)
			}
		}
	}()
}

func FetchChartDataFromDB(coinID string) ([]types.ChartData, error) {
	conn := db.GetConnection()

	var chartData []types.ChartData
	if err := conn.Model(&model.ChartData{}).
		Select("open_time, close_time, open, high, low, close, volume").
		Where("symbol = ?", coinID).
		Order("close_time desc").
		Find(&chartData).Error; err != nil {
		log.Println(err)
		return nil, err
	}
	return chartData, nil
}
