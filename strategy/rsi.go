package strategy

import (
	"errors"
)

// CalculateRSI computes the Relative Strength Index.
func CalculateRSI(prices []float64, period int) (float64, error) {
	if len(prices) < period+1 {
		return 0, errors.New("not enough data to calculate RSI")
	}

	var totalGain, totalLoss float64

	for i := 1; i <= period; i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			totalGain += change
		} else {
			totalLoss -= change 
		}
	}

	avgGain := totalGain / float64(period)
	avgLoss := totalLoss / float64(period)

	for i := period + 1; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		var currentGain, currentLoss float64

		if change > 0 {
			currentGain = change
		} else {
			currentLoss = -change
		}

		avgGain = (avgGain*float64(period-1) + currentGain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + currentLoss) / float64(period)
	}

	if avgLoss == 0 {
		return 100.0, nil 
	}

	rs := avgGain / avgLoss
	rsi := 100.0 - (100.0 / (1.0 + rs))

	return rsi, nil
}
