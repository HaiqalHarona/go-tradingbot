package main

import (
	"log"
	"time"

	"tradingbot/config"
	"tradingbot/data"
	"tradingbot/execution"
	"tradingbot/strategy"
)

func main() {
	// Configure logger with date, time, and microsecond precision timestamps
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	// Initialize configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	log.Println("config fetched successfully")

	// Init risk management
	rg := execution.NewRiskGuard(cfg)

	// define stock pool
	stockPool := cfg.StockPool
	barsRequired := 50

	// init trading loop
	for {
		log.Println("Starting trading loop")

		// Check stock pool for buy signals
		for _, ticker := range stockPool {
			prices, err := data.FetchClosingPrices(ticker, barsRequired)
			if err != nil {
				log.Printf("[%s] Data fetch error skipped: %v\n", ticker, err)
				continue
			}

			// Run prices through the indicators
			signal, err := strategy.EvaluateStrategy(prices)
			if err != nil {
				log.Printf("[%s] Strategy processing failed: %v\n", ticker, err)
				continue
			}

			// Execute orders
			if signal == strategy.SignalBuy {
				currentPrice := prices[len(prices)-1]
				log.Printf("[%s] 🟢 BUY SIGNAL HIT: Stock %s triggered buy signal. Routing fractional order...\n", ticker, ticker)
				if err := rg.ExecuteFractionalBuy(ticker, currentPrice); err != nil {
					log.Printf("[%s] ❌ Order routing failed for stock %s: %v\n", ticker, ticker, err)
				}
			} else {
				log.Printf("[%s] ⚪ Market State: HOLD\n", ticker)
			}
		}

		log.Println("Scan complete, starting for next cycle in 1 hour")
		time.Sleep(1 * time.Hour)
	}
}