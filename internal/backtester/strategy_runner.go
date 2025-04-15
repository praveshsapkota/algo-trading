package backtester

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/user/algo-trading/internal/signal-generator"
)

// BacktestContext holds the context for a backtest run
type BacktestContext struct {
	Symbol       string
	Data         []HistoricalData
	CurrentIndex int
	Trades       []BacktestTrade
	Cash         float64
	Positions    map[string]float64
	Parameters   map[string]interface{}
}

// RunBacktest executes a backtest for a given strategy and parameters
func (bs *BacktesterService) RunBacktest(ctx context.Context, request *BacktestRequest, strategy signalgenerator.TradingStrategy) (*BacktestResult, error) {
	// Validate request
	if len(request.Symbols) == 0 {
		return nil, fmt.Errorf("at least one symbol is required")
	}

	// Initialize result
	result := &BacktestResult{
		StrategyName: request.StrategyName,
		Symbols:      request.Symbols,
		StartDate:    request.StartDate,
		EndDate:      request.EndDate,
		Parameters:   request.Parameters,
		CreatedAt:    time.Now(),
	}

	// Run backtest for each symbol
	var allTrades []BacktestTrade
	initialCash := 100000.0 // $100,000 starting capital
	totalPnL := 0.0

	for _, symbol := range request.Symbols {
		// Fetch historical data
		data, err := bs.GetHistoricalData(ctx, symbol, request.Timeframe, request.StartDate, request.EndDate)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch historical data for %s: %v", symbol, err)
		}

		if len(data) == 0 {
			return nil, fmt.Errorf("no historical data found for %s", symbol)
		}

		// Sort data by timestamp
		sort.Slice(data, func(i, j int) bool {
			return data[i].Timestamp.Before(data[j].Timestamp)
		})

		// Create backtest context
		btContext := &BacktestContext{
			Symbol:     symbol,
			Data:       data,
			Cash:       initialCash,
			Positions:  make(map[string]float64),
			Parameters: request.Parameters,
		}

		// Run strategy simulation
		trades, err := bs.simulateStrategy(ctx, btContext, strategy)
		if err != nil {
			return nil, fmt.Errorf("strategy simulation failed: %v", err)
		}

		// Calculate PnL for this symbol
		symbolPnL := 0.0
		for _, trade := range trades {
			symbolPnL += trade.PnL
		}
		totalPnL += symbolPnL

		// Add trades to result
		allTrades = append(allTrades, trades...)
	}

	// Calculate performance metrics
	result.TotalReturn = (totalPnL / initialCash) * 100
	result.TradeCount = len(allTrades)
	result.Trades = allTrades

	// Calculate win rate
	winCount := 0
	for _, trade := range allTrades {
		if trade.PnL > 0 {
			winCount++
		}
	}
	if result.TradeCount > 0 {
		result.WinRate = float64(winCount) / float64(result.TradeCount) * 100
	}

	// Calculate max drawdown
	result.MaxDrawdown = bs.calculateMaxDrawdown(allTrades, initialCash)

	// Calculate Sharpe ratio
	result.SharpeRatio = bs.calculateSharpeRatio(allTrades)

	// Save result to database
	if err := bs.db.Create(result).Error; err != nil {
		return nil, fmt.Errorf("failed to save backtest result: %v", err)
	}

	return result, nil
}

// simulateStrategy simulates a trading strategy on historical data
func (bs *BacktesterService) simulateStrategy(ctx context.Context, btContext *BacktestContext, strategy signalgenerator.TradingStrategy) ([]BacktestTrade, error) {
	var trades []BacktestTrade
	var openPosition *BacktestTrade
	data := btContext.Data

	// Process each data point
	for i := 0; i < len(data); i++ {
		// Skip if not enough data for lookback
		if i < 20 { // Minimum lookback for most strategies
			continue
		}

		// Create a slice of recent data for the strategy
		lookback := 50 // Lookback period
		if i < lookback {
			lookback = i
		}
		recentData := data[i-lookback : i+1]

		// Create a mock quote for the strategy
		currentBar := data[i]
		mockQuote := createMockQuote(btContext.Symbol, currentBar)

		// Generate signal
		signal, err := strategy.GenerateSignal(mockQuote)
		if err != nil {
			bs.logger.Warnf("Failed to generate signal at %s: %v", currentBar.Timestamp, err)
			continue
		}

		// Process signal
		if signal != nil {
			// Handle buy signal
			if signal.Type == signalgenerator.BuySignal && openPosition == nil {
				// Calculate position size (simple implementation)
				cash := btContext.Cash
				price := currentBar.Close
				quantity := (cash * 0.02) / price // Risk 2% per trade
				if quantity <= 0 {
					continue
				}

				// Create new trade
				openPosition = &BacktestTrade{
					Symbol:     btContext.Symbol,
					Type:       string(signal.Type),
					EntryPrice: price,
					Quantity:   quantity,
					EntryTime:  currentBar.Timestamp,
					Strategy:   signal.Strategy,
				}

				// Update cash and positions
				btContext.Cash -= price * quantity
				btContext.Positions[btContext.Symbol] = quantity
			}

			// Handle sell signal
			if signal.Type == signalgenerator.SellSignal && openPosition != nil {
				// Close position
				price := currentBar.Close
				quantity := btContext.Positions[btContext.Symbol]
				pnl := (price - openPosition.EntryPrice) * quantity

				// Update trade
				openPosition.ExitPrice = price
				openPosition.ExitTime = currentBar.Timestamp
				openPosition.PnL = pnl

				// Add to trades list
				trades = append(trades, *openPosition)

				// Update cash and positions
				btContext.Cash += price * quantity
				delete(btContext.Positions, btContext.Symbol)
				openPosition = nil
			}
		}

		// Check for stop loss or take profit (simple implementation)
		if openPosition != nil {
			price := currentBar.Close
			quantity := btContext.Positions[btContext.Symbol]
			unrealizedPnL := (price - openPosition.EntryPrice) * quantity
			stopLossPercent := -0.02 // 2% stop loss
			takeProfitPercent := 0.04 // 4% take profit

			// Check stop loss
			if unrealizedPnL / (openPosition.EntryPrice * quantity) <= stopLossPercent {
				// Close position with stop loss
				openPosition.ExitPrice = price
				openPosition.ExitTime = currentBar.Timestamp
				openPosition.PnL = unrealizedPnL
				trades = append(trades, *openPosition)

				// Update cash and positions
				btContext.Cash += price * quantity
				delete(btContext.Positions, btContext.Symbol)
				openPosition = nil
			}

			// Check take profit
			if unrealizedPnL / (openPosition.EntryPrice * quantity) >= takeProfitPercent {
				// Close position with take profit
				openPosition.ExitPrice = price
				openPosition.ExitTime = currentBar.Timestamp
				openPosition.PnL = unrealizedPnL
				trades = append(trades, *openPosition)

				// Update cash and positions
				btContext.Cash += price * quantity
				delete(btContext.Positions, btContext.Symbol)
				openPosition = nil
			}
		}
	}

	// Close any remaining open positions at the last price
	if openPosition != nil {
		lastBar := data[len(data)-1]
		price := lastBar.Close
		quantity := btContext.Positions[btContext.Symbol]
		pnl := (price - openPosition.EntryPrice) * quantity

		// Update trade
		openPosition.ExitPrice = price
		openPosition.ExitTime = lastBar.Timestamp
		openPosition.PnL = pnl

		// Add to trades list
		trades = append(trades, *openPosition)
	}

	return trades, nil
}

// createMockQuote creates a mock quote for the strategy
func createMockQuote(symbol string, bar HistoricalData) *alphavantage.QuoteResponse {
	return &alphavantage.QuoteResponse{
		GlobalQuote: alphavantage.GlobalQuote{
			Symbol:        symbol,
			Price:         fmt.Sprintf("%.2f", bar.Close),
			Volume:        fmt.Sprintf("%d", bar.Volume),
			LatestDay:     bar.Timestamp.Format("2006-01-02"),
			PreviousClose: fmt.Sprintf("%.2f", bar.Open),
		},
	}
}

// calculateMaxDrawdown calculates the maximum drawdown from a series of trades
func (bs *BacktesterService) calculateMaxDrawdown(trades []BacktestTrade, initialCapital float64) float64 {
	if len(trades) == 0 {
		return 0
	}

	// Sort trades by entry time
	sort.Slice(trades, func(i, j int) bool {
		return trades[i].EntryTime.Before(trades[j].EntryTime)
	})

	// Calculate equity curve
	equity := initialCapital
	peak := equity
	maxDrawdown := 0.0

	for _, trade := range trades {
		equity += trade.PnL
		if equity > peak {
			peak = equity
		}
		drawdown := (peak - equity) / peak * 100
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown
}

// calculateSharpeRatio calculates the Sharpe ratio from a series of trades
func (bs *BacktesterService) calculateSharpeRatio(trades []BacktestTrade) float64 {
	if len(trades) < 2 {
		return 0
	}

	// Calculate daily returns
	dailyReturns := make(map[string]float64)
	for _, trade := range trades {
		date := trade.ExitTime.Format("2006-01-02")
		dailyReturns[date] += trade.PnL
	}

	// Convert to slice for calculations
	var returns []float64
	for _, ret := range dailyReturns {
		returns = append(returns, ret)
	}

	// Calculate mean return
	sum := 0.0
	for _, ret := range returns {
		sum += ret
	}
	meanReturn := sum / float64(len(returns))

	// Calculate standard deviation
	sumSquaredDiff := 0.0
	for _, ret := range returns {
		diff := ret - meanReturn
		sumSquaredDiff += diff * diff
	}
	stdDev := math.Sqrt(sumSquaredDiff / float64(len(returns)))

	// Calculate annualized Sharpe ratio (assuming 252 trading days per year)
	if stdDev == 0 {
		return 0
	}
	sharpeRatio := (meanReturn * 252) / (stdDev * math.Sqrt(252))

	return sharpeRatio
}