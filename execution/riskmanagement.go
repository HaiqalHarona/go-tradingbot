package execution

import (
	"fmt"
	"log"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/shopspring/decimal"
)

// RiskGuard evaluates portfolio positions and manages trade execution risk.
// It wraps the Alpaca Trade API client to interact with account details and orders.
type RiskGuard struct {
	client *alpaca.Client
}

// NewRiskGuard instantiates RiskGuard using environment variables for credentials.
// Under the hood, alpaca.NewClient(alpaca.ClientOpts{}) automatically searches for
// APCA_API_KEY_ID, APCA_API_SECRET_KEY, and APCA_API_BASE_URL env vars.
func NewRiskGuard() *RiskGuard {
	return &RiskGuard{
		client: alpaca.NewClient(alpaca.ClientOpts{}),
	}
}


// ExecuteFractionalBuy places a market buy order allocating 2% of equity with a 1.5% stop loss attached.
// It also enforces a maximum limit of 4 active open positions/orders.
func (rg *RiskGuard) ExecuteFractionalBuy(ticker string, currentPrice float64) error {
	// Fetch account info to calculate allocation and verify buying power
	account, err := rg.client.GetAccount()
	if err != nil {
		return fmt.Errorf("failed to fetch account: %w", err)
	}

	// Calculate a dollar allocation equal to 2% of the account's total equity
	equity := account.Equity.InexactFloat64()
	allocationAmount := equity * 0.02

	// Verify that we have sufficient buying power to cover the allocation
	buyingPower := account.BuyingPower.InexactFloat64()
	if buyingPower < allocationAmount {
		return fmt.Errorf("insufficient buying power: need $%.2f, have $%.2f", allocationAmount, buyingPower)
	}

	// Retrieve open positions to check if we already own this ticker
	positions, err := rg.client.GetPositions()
	if err != nil {
		return fmt.Errorf("failed to fetch active positions: %w", err)
	}

	for _, pos := range positions {
		if pos.Symbol == ticker {
			log.Printf("[%s] Position already open for stock %s. Skipping trade.\n", ticker, ticker)
			return nil
		}
	}

	// Retrieve open orders to check overall active position + order cap
	openOrders, _ := rg.client.GetOrders(alpaca.GetOrdersRequest{Status: "open"})
	if len(positions)+len(openOrders) >= 4 {
		log.Printf("[LIMIT REACHED] Maximum limit of 4 active open positions/orders reached. Skipping buy for stock %s.\n", ticker)
		return nil
	}

	// Convert float64 values to decimal.Decimal required by Alpaca
	dollarAmountDecimal := decimal.NewFromFloat(allocationAmount)
	stopPriceDecimal := decimal.NewFromFloat(currentPrice * 0.985)   // 1.5% stop loss
	takeProfitDecimal := decimal.NewFromFloat(currentPrice * 1.030)  // 3.0% take profit (2:1 risk/reward)

	// PlaceOrderRequest configures our trade with attached bracket TakeProfit & StopLoss
	req := alpaca.PlaceOrderRequest{
		Symbol:      ticker,
		Notional:    &dollarAmountDecimal, // Dollar amount to invest
		Side:        alpaca.Buy,           // Buy side order
		Type:        alpaca.Market,        // Market price order
		TimeInForce: alpaca.GTC,           // Good 'Til Canceled for order & bracket legs
		OrderClass:  alpaca.Bracket,       // Attach bracket orders
		TakeProfit: &alpaca.TakeProfit{
			LimitPrice: &takeProfitDecimal,
		},
		StopLoss: &alpaca.StopLoss{
			StopPrice: &stopPriceDecimal,
		},
	}

	// PlaceOrder sends a POST request to /v2/orders to execute the trade
	order, err := rg.client.PlaceOrder(req)
	if err != nil {
		return fmt.Errorf("failed to place order for %s: %w", ticker, err)
	}

	log.Printf("[ORDER PLACED] Allocated $%.2f to stock %s with 3.0%% Take-Profit ($%.2f) & 1.5%% Stop-Loss ($%.2f). Order ID: %s\n", allocationAmount, ticker, currentPrice*1.030, currentPrice*0.985, order.ID)
	return nil
}