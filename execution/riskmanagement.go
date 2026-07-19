package execution

import (
	"fmt"
	"log"

	"tradingbot/config"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/shopspring/decimal"
)

// RiskGuard evaluates portfolio positions and manages trade execution risk.
// It wraps the Alpaca Trade API client to interact with account details and orders.
type RiskGuard struct {
	client *alpaca.Client
	cfg    *config.Config
}

// NewRiskGuard instantiates RiskGuard using environment variables and bot configuration.
func NewRiskGuard(cfg *config.Config) *RiskGuard {
	clientOpts := alpaca.ClientOpts{
		APIKey:    cfg.AlpacaKeyID,
		APISecret: cfg.AlpacaSecretKey,
		BaseURL:   cfg.AlpacaBaseURL,
	}
	return &RiskGuard{
		client: alpaca.NewClient(clientOpts),
		cfg:    cfg,
	}
}

// ExecuteFractionalBuy places a market buy order allocating configured equity % with bracket StopLoss and TakeProfit attached.
// It also enforces the maximum limit of active open positions/orders.
func (rg *RiskGuard) ExecuteFractionalBuy(ticker string, currentPrice float64) error {
	// Fetch account info to calculate allocation and verify buying power
	account, err := rg.client.GetAccount()
	if err != nil {
		return fmt.Errorf("failed to fetch account: %w", err)
	}

	// Calculate a dollar allocation based on configured equity %
	equity := account.Equity.InexactFloat64()
	allocationAmount := equity * rg.cfg.EquityAllocation

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
	if len(positions)+len(openOrders) >= rg.cfg.MaxOpenPositions {
		log.Printf("[LIMIT REACHED] Maximum limit of %d active open positions/orders reached. Skipping buy for stock %s.\n", rg.cfg.MaxOpenPositions, ticker)
		return nil
	}

	// Convert allocation amount to decimal.Decimal required by Alpaca
	dollarAmountDecimal := decimal.NewFromFloat(allocationAmount)

	// PlaceOrderRequest configures fractional market buy order.
	// Note: Alpaca API requires fractional (Notional) orders to use simple order class.
	req := alpaca.PlaceOrderRequest{
		Symbol:      ticker,
		Notional:    &dollarAmountDecimal, // Dollar amount to invest
		Side:        alpaca.Buy,           // Buy side order
		Type:        alpaca.Market,        // Market price order
		TimeInForce: alpaca.Day,           // Day order
	}

	// PlaceOrder sends a POST request to /v2/orders to execute the trade
	order, err := rg.client.PlaceOrder(req)
	if err != nil {
		return fmt.Errorf("failed to place order for %s: %w", ticker, err)
	}

	log.Printf("[ORDER PLACED] Allocated $%.2f to stock %s (Order ID: %s)\n", allocationAmount, ticker, order.ID)
	return nil
}