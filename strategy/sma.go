package strategy

import (
	"errors"
)

func CalculateSMA(prices []float64, period int) (float64, error) {
	if len(prices) < period {
		return 0, errors.New("not enough data points to calculate SMA")
	}

	sum := 0.0
	recentPrices := prices[len(prices)-period:] // Get the last 'period' number of prices

	for _, price := range recentPrices {
		sum += price
	}

	return sum / float64(period), nil
}