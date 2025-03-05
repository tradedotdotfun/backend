package binance

import (
	"log"
	"time"

	"tradedotdotfun-backend/indexer/config"
)

type Fetcher struct {
	chartFetcher       *ChartFetcher
	priceFetcher       *PriceFetcher
	chartFetcherTicker *time.Ticker
	priceFetcherTicker *time.Ticker
}

func NewFetcher() *Fetcher {
	return &Fetcher{
		chartFetcher:       NewChartFetcher(),
		priceFetcher:       NewPriceFetcher(),
		chartFetcherTicker: time.NewTicker(config.CHART_UPDATE_INTERVAL),
		priceFetcherTicker: time.NewTicker(config.PRICE_UPDATE_INTERVAL),
	}
}

func (e *Fetcher) Fetch() {
	e.runChartUpdateIndexer()
	e.runPriceUpdateIndexer()
}

func (e *Fetcher) runChartUpdateIndexer() {
	log.Println("Run ChartFetcher")
	start := time.Now()
	e.chartFetcher.Fetch()
	log.Println("Run ChartFetcher Completed, elapsed:", time.Since(start))
	go func() {
		for range e.chartFetcherTicker.C {
			log.Println("Run ChartFetcher")
			start := time.Now()
			e.chartFetcher.Fetch()
			log.Println("Run ChartFetcher Completed, elapsed:", time.Since(start))
		}
	}()
}

func (e *Fetcher) runPriceUpdateIndexer() {
	e.priceFetcher.Fetch()
	go func() {
		for range e.priceFetcherTicker.C {
			e.priceFetcher.Fetch()
		}
	}()
}
