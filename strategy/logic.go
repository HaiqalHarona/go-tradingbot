package strategy

import (
	"fmt"
)

type TradeSignal string

const (
	SignalBuy  TradeSignal = "BUY"
	SignalHold TradeSignal = "HOLD"
)

// EvaluateStrategy analyzes the price history to determine if a buying window is open.
func EvaluateStrategy(prices []float64) (TradeSignal, error) {
	const smaPeriod = 50
	const rsiPeriod = 14

	// Calculate the 50-period SMA
	sma, err := CalculateSMA(prices, smaPeriod)
	if err != nil {
		return SignalHold, fmt.Errorf("SMA calculation failed: %w", err)
	}

	// Calculate the 14-period RSI
	rsi, err := CalculateRSI(prices, rsiPeriod)
	if err != nil {
		return SignalHold, fmt.Errorf("RSI calculation failed: %w", err)
	}

	if len(prices) == 0 {
		return SignalHold, fmt.Errorf("cannot evaluate empty price slice")
	}
	currentPrice := prices[len(prices)-1]

	// The Buy Trigger:
	// Price must be above the 50 SMA (Upward Trend) AND RSI must be below 30 (Oversold Dip)
	if currentPrice > sma && rsi < 30 {
		return SignalBuy, nil
	}

	return SignalHold, nil
}
