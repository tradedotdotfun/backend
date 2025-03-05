package cache

var round uint64 = 0
var priceMap = make(map[string]float64)

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
