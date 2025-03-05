package binance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"tradedotdotfun-backend/common/model"
)

const (
	KILINES_API_URL  = "https://api.binance.com/api/v3/klines"
	PRICE_API_URL    = "https://api.binance.com/api/v3/ticker/price"
	KILINES_LIMIT    = 1000
	KILINES_INTERVAL = "1h"
)

type BinanceCandle []interface{}

func GetChartData(symbol string, lastTime, endTime int64) ([]model.ChartData, error) {
	startTime := lastTime + 1
	var allCandles []model.ChartData
	url := fmt.Sprintf("%s?symbol=%s&interval=%s&startTime=%d&limit=%d", KILINES_API_URL, symbol, KILINES_INTERVAL, startTime, KILINES_LIMIT)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data []BinanceCandle
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	var candles []model.ChartData
	for _, c := range data {
		openTime := time.Unix(0, int64(c[0].(float64))*int64(time.Millisecond)).UTC()
		closeTime := time.Unix(0, int64(c[6].(float64))*int64(time.Millisecond)).UTC()
		candles = append(candles, model.ChartData{
			Symbol:    symbol,
			OpenTime:  openTime,
			CloseTime: closeTime,
			Open:      parseFloat(c[1]),
			High:      parseFloat(c[2]),
			Low:       parseFloat(c[3]),
			Close:     parseFloat(c[4]),
			Volume:    parseFloat(c[5]),
		})
	}

	allCandles = append(allCandles, candles...)
	time.Sleep(time.Millisecond * 500) // 요청 제한 방지

	return allCandles, nil
}

func GetTokenPrices(symbols []string) (map[string]float64, error) {
	prices := make(map[string]float64)
	for _, symbol := range symbols {
		url := fmt.Sprintf("%s?symbol=%s", PRICE_API_URL, symbol)
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var result struct {
			Symbol string `json:"symbol"`
			Price  string `json:"price"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		prices[result.Symbol] = parseFloat(result.Price)
		time.Sleep(time.Millisecond * 100) // 요청 제한 방지
	}
	return prices, nil
}

func parseFloat(value interface{}) float64 {
	if v, ok := value.(string); ok {
		var result float64
		fmt.Sscanf(v, "%f", &result)
		return result
	}
	return 0
}
