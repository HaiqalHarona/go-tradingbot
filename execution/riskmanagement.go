package execution

import (
	"fmt"

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

// EnforceStopLoss closes positions whose unrealized loss exceeds 1.5% of total account equity.
// Flow:
// 1. Fetch account info to get the total equity value.
// 2. Compute the 1.5% maximum dollar loss threshold based on that equity.
// 3. Retrieve all open positions from Alpaca.
// 4. Iterate over each position, checking the absolute unrealized PnL in dollars.
// 5. If any position's loss exceeds the threshold, call ClosePosition to liquidate it immediately.
func (rg *RiskGuard) EnforceStopLoss() error {
	// Call GetAccount to retrieve account state (balance, equity, buying power, etc.)
	account, err := rg.client.GetAccount()
	if err != nil {
		return fmt.Errorf("failed to fetch account: %w", err)
	}
	
	// Convert account equity (which is an Alpaca decimal) to float64 for easy math
	equity := account.Equity.InexactFloat64()
	
	// Calculate the stop-loss limit in dollars (1.5% of total equity)
	maxLoss := equity * 0.015

	// Call GetPositions to fetch a slice of all active holdings in the account
	positions, err := rg.client.GetPositions()
	if err != nil {
		return fmt.Errorf("failed to fetch active positions: %w", err)
	}

	for _, pos := range positions {
		// UnrealizedPL represents the current PnL in dollars. If nil, skip.
		if pos.UnrealizedPL == nil {
			continue
		}
		plDollar := pos.UnrealizedPL.InexactFloat64()

		// If plDollar is negative and its absolute value is larger than maxLoss, trigger stop-loss
		if plDollar <= -maxLoss {
			fmt.Printf("[RISK ALERT] %s loss ($%.2f) exceeds 1.5%% of total equity ($%.2f). Executing market exit.\n", pos.Symbol, plDollar, maxLoss)
			
			// ClosePosition sends a DELETE request to /v2/positions/{symbol} to liquidate the asset.
			// Passing an empty ClosePositionRequest tells Alpaca to liquidate 100% of the holding.
			if _, err := rg.client.ClosePosition(pos.Symbol, alpaca.ClosePositionRequest{}); err != nil {
				fmt.Printf("[ERROR] Stop-loss close failed for %s: %v\n", pos.Symbol, err)
			}
		}
	}
	return nil
}

// ExecuteFractionalBuy places a market buy order allocating 2% of equity.
// Flow:
// 1. Fetch account info to calculate the 2% allocation from total equity.
// 2. Verify if the account has enough actual buying power to afford the purchase.
// 3. Scan existing positions to ensure we don't buy an asset we already own.
// 4. Construct a fractional PlaceOrderRequest using the "Notional" field.
// 5. Submit the order to Alpaca.
func (rg *RiskGuard) ExecuteFractionalBuy(ticker string) error {
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
	positions, _ := rg.client.GetPositions()
	for _, pos := range positions {
		if pos.Symbol == ticker {
			fmt.Printf("[%s] Position already open. Skipping trade.\n", ticker)
			return nil
		}
	}

	// Convert our float64 allocation amount to a decimal.Decimal type required by Alpaca
	dollarAmountDecimal := decimal.NewFromFloat(allocationAmount)

	// PlaceOrderRequest configures our trade.
	// By using "Notional" instead of "Qty", we specify a dollar budget for the trade,
	// allowing Alpaca to purchase fractional shares if the stock price is higher than our budget.
	req := alpaca.PlaceOrderRequest{
		Symbol:      ticker,
		Notional:    &dollarAmountDecimal, // Dollar amount to invest
		Side:        alpaca.Buy,           // Buy side order
		Type:        alpaca.Market,        // Market price order
		TimeInForce: alpaca.Day,           // Valid for the current trading day
	}

	// PlaceOrder sends a POST request to /v2/orders to execute the trade
	order, err := rg.client.PlaceOrder(req)
	if err != nil {
		return fmt.Errorf("failed to place order for %s: %w", ticker, err)
	}

	fmt.Printf("[ORDER PLACED] Allocated $%.2f to %s. Order ID: %s\n", allocationAmount, ticker, order.ID)
	return nil
}