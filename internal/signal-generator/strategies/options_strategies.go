package strategies

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/user/algo-trading/internal/signal-generator"
	"github.com/user/algo-trading/pkg/api/alphavantage"
	"github.com/user/algo-trading/pkg/models"
	"github.com/user/algo-trading/pkg/options"
)

// SimpleOptionsStrategy implements a basic options trading strategy
type SimpleOptionsStrategy struct {
	avClient       *alphavantage.Client
	pricingModel   *options.BlackScholesModel
	riskFreeRate   float64
	volatilityLookback int
	minVolatilityThreshold float64
	maxVolatilityThreshold float64
}

// NewSimpleOptionsStrategy creates a new simple options strategy
func NewSimpleOptionsStrategy(
	client *alphavantage.Client,
	riskFreeRate float64,
	volatilityLookback int,
	minVolatilityThreshold float64,
	maxVolatilityThreshold float64,
) *SimpleOptionsStrategy {
	return &SimpleOptionsStrategy{
		avClient:       client,
		pricingModel:   options.NewBlackScholesModel(),
		riskFreeRate:   riskFreeRate,
		volatilityLookback: volatilityLookback,
		minVolatilityThreshold: minVolatilityThreshold,
		maxVolatilityThreshold: maxVolatilityThreshold,
	}
}

// GetName returns the strategy name
func (s *SimpleOptionsStrategy) GetName() string {
	return fmt.Sprintf("SIMPLE_OPTIONS_%d", s.volatilityLookback)
}

// GenerateSignal implements the TradingStrategy interface
func (s *SimpleOptionsStrategy) GenerateSignal(quote *alphavantage.QuoteResponse) (*signalgenerator.TradingSignal, error) {
	// Get current price
	price, err := strconv.ParseFloat(quote.GlobalQuote.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse current price: %v", err)
	}

	// Get historical volatility
	volatility, trend, err := s.calculateHistoricalVolatility(quote.GlobalQuote.Symbol, s.volatilityLookback)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate historical volatility: %v", err)
	}

	// Check if volatility is within thresholds
	if volatility < s.minVolatilityThreshold || volatility > s.maxVolatilityThreshold {
		return nil, nil // No signal if volatility is outside our thresholds
	}

	// Generate signal based on volatility and trend
	var signal *signalgenerator.TradingSignal
	var optionType models.OptionType
	var confidence float64

	if trend > 0 {
		// Uptrend - consider buying calls
		optionType = models.Call
		confidence = trend * volatility * 100 // Simple confidence calculation
		
		signal = &signalgenerator.TradingSignal{
			Symbol:     quote.GlobalQuote.Symbol,
			Type:       signalgenerator.BuySignal,
			Price:      price,
			Timestamp:  time.Now(),
			Strategy:   s.GetName(),
			Confidence: confidence,
			Metadata: map[string]interface{}{
				"option_type": string(optionType),
				"volatility":  volatility,
				"trend":       trend,
			},
		}
	} else if trend < 0 {
		// Downtrend - consider buying puts
		optionType = models.Put
		confidence = -trend * volatility * 100 // Simple confidence calculation
		
		signal = &signalgenerator.TradingSignal{
			Symbol:     quote.GlobalQuote.Symbol,
			Type:       signalgenerator.BuySignal,
			Price:      price,
			Timestamp:  time.Now(),
			Strategy:   s.GetName(),
			Confidence: confidence,
			Metadata: map[string]interface{}{
				"option_type": string(optionType),
				"volatility":  volatility,
				"trend":       trend,
			},
		}
	}

	return signal, nil
}

// calculateHistoricalVolatility calculates the historical volatility and trend
func (s *SimpleOptionsStrategy) calculateHistoricalVolatility(symbol string, lookback int) (float64, float64, error) {
	// Get daily time series
	timeSeries, err := s.avClient.GetTimeSeries(symbol, "daily", "compact")
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get time series data: %v", err)
	}

	// Extract closing prices
	var prices []float64
	var dates []string

	for date, values := range timeSeries.TimeSeries {
		closePrice, err := strconv.ParseFloat(values.Close, 64)
		if err != nil {
			continue
		}
		prices = append(prices, closePrice)
		dates = append(dates, date)
	}

	// Sort prices by date (most recent first)
	sortPricesByDate(prices, dates)

	// Ensure we have enough data
	if len(prices) < lookback+1 {
		return 0, 0, fmt.Errorf("not enough historical data, need at least %d days", lookback+1)
	}

	// Calculate daily returns
	returns := make([]float64, len(prices)-1)
	for i := 0; i < len(prices)-1; i++ {
		returns[i] = (prices[i] / prices[i+1]) - 1
	}

	// Calculate volatility (standard deviation of returns)
	volatility := calculateStandardDeviation(returns[:lookback])
	
	// Annualize volatility (assuming 252 trading days in a year)
	annualizedVolatility := volatility * math.Sqrt(252)

	// Calculate trend (simple moving average direction)
	trend := calculateTrend(prices, lookback)

	return annualizedVolatility, trend, nil
}

// sortPricesByDate sorts prices by date (most recent first)
func sortPricesByDate(prices []float64, dates []string) {
	// This is a simplified implementation
	// In a real system, we would sort based on actual dates
	// For now, we assume the data is already sorted by date
}

// calculateStandardDeviation calculates the standard deviation of a slice of values
func calculateStandardDeviation(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// Calculate sum of squared differences
	sumSquaredDiff := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquaredDiff += diff * diff
	}

	// Calculate standard deviation
	return math.Sqrt(sumSquaredDiff / float64(len(values)))
}

// calculateTrend calculates the trend direction based on simple moving averages
func calculateTrend(prices []float64, lookback int) float64 {
	if len(prices) < lookback*2 {
		return 0
	}

	// Calculate short-term SMA (5 days)
	shortTermPeriod := 5
	shortTermSMA := calculateSMA(prices[:shortTermPeriod])

	// Calculate medium-term SMA (lookback days)
	mediumTermSMA := calculateSMA(prices[:lookback])

	// Calculate trend strength (-1 to 1)
	if mediumTermSMA == 0 {
		return 0
	}
	
	trendStrength := (shortTermSMA - mediumTermSMA) / mediumTermSMA
	
	// Normalize trend strength to be between -1 and 1
	if trendStrength > 1 {
		trendStrength = 1
	} else if trendStrength < -1 {
		trendStrength = -1
	}
	
	return trendStrength
}

// calculateSMA calculates the simple moving average of a slice of prices
func calculateSMA(prices []float64) float64 {
	if len(prices) == 0 {
		return 0
	}

	sum := 0.0
	for _, price := range prices {
		sum += price
	}
	return sum / float64(len(prices))
}