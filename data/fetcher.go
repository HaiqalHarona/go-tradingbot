package data

import (
	"fmt"
	"time"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
)

// Fetch closing prices last 'n' days of daily closing prices for a given ticker
func FetchClosingPrices(ticker string, barsReq int) ([]float64, error) {

	// Initialize the market data client. Zero-config uses APCA_API_KEY_ID and APCA_API_SECRET_KEY env vars.
	client := marketdata.NewClient(marketdata.ClientOpts{})

	loopback := (barsReq / 7) + 10
	startTime := time.Now().AddDate(0, 0, -loopback) // Calculate the start time for fetching bars, going back 'loopback' days.

	// Define query parameters for historical bars. 
	// TimeFrame defines bar interval, Feed specifies data source (IEX or SIP).
	req := marketdata.GetBarsRequest{
		TimeFrame: marketdata.OneHour,
		Start:     startTime,
		End:       time.Now(),
		Feed:      marketdata.IEX,
	}

	// Fetch historical bar data for the specified symbol using the Alpaca client.
	bars, err := client.GetBars(ticker, req)
	if err != nil {
		return nil, fmt.Errorf("error fetching bars for %s: %v", ticker, err)
	}

	// Since we queried a larger window to account for weekends/holidays, 
	// we slice the last 'days' elements to get the most recent trading days.
	if len(bars) < barsReq {
		return nil, fmt.Errorf("not enough historical data returned for %s: got %d, need %d", ticker, len(bars), barsReq)
	}
	recentBars := bars[len(bars)-barsReq:]

	var prices []float64
	for _, bar := range recentBars {
		// Extract the closing price from each returned Alpaca Bar struct.
		prices = append(prices, bar.Close)
	}

	return prices, nil
}