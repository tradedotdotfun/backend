package okx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"tradedotdotfun-backend/common/model"
)

const (
	baseURL = "https://www.okx.com"
)

var (
	OkxSymbolMap = map[string]string{
		"BTCUSDT":   "BTC-USDT",
		"ETHUSDT":   "ETH-USDT",
		"SOLUSDT":   "SOL-USDT",
		"SONICUSDT": "SONIC-USDT",
	}
)

// GetCandles fetches candlestick data from OKX API
func GetChartData(symbol string, before time.Time) ([]model.ChartData, error) {
	url := fmt.Sprintf("%s/api/v5/market/history-candles?instId=%s&bar=1H&before=%s", baseURL, OkxSymbolMap[symbol], strconv.FormatInt(before.UnixMilli(), 10))
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch data: %s", resp.Status)
	}

	var data struct {
		Code string     `json:"code"`
		Data [][]string `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if data.Code != "0" {
		return nil, fmt.Errorf("API error: %s", data.Code)
	}

	if len(data.Data) == 0 {
		return nil, nil
	}

	var candles []model.ChartData
	for _, d := range data.Data {
		timestamp, _ := strconv.ParseInt(d[0], 10, 64)
		open, _ := strconv.ParseFloat(d[1], 64)
		high, _ := strconv.ParseFloat(d[2], 64)
		low, _ := strconv.ParseFloat(d[3], 64)
		close, _ := strconv.ParseFloat(d[4], 64)
		volume, _ := strconv.ParseFloat(d[5], 64)
		candles = append(candles, model.ChartData{
			Symbol:    symbol,
			OpenTime:  time.UnixMilli(timestamp).UTC(),
			CloseTime: time.UnixMilli(timestamp + 3600*1000 - 1).UTC(),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
		})
	}
	time.Sleep(time.Millisecond * 500) // 요청 제한 방지

	return candles, nil
}

func GetTokenPrices(symbols []string) (map[string]float64, error) {
	url := fmt.Sprintf("%s/api/v5/market/tickers?instType=SPOT", baseURL)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch data: %s", resp.Status)
	}

	var data struct {
		Code string `json:"code"`
		Data []struct {
			InstId string `json:"instId"`
			Last   string `json:"last"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if data.Code != "0" {
		return nil, fmt.Errorf("API error: %s", data.Code)
	}

	prices := make(map[string]float64)
	for _, d := range data.Data {
		for _, symbol := range symbols {
			okxSymbol := OkxSymbolMap[symbol]
			if d.InstId == okxSymbol {
				price, _ := strconv.ParseFloat(d.Last, 64)
				prices[symbol] = price
			}
		}
	}
	time.Sleep(time.Millisecond * 100) // 요청 제한 방지

	return prices, nil
}
