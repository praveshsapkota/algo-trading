package strategies

import (
	"fmt"
	"strconv"
	"time"

	"algo-trading/internal/signal-generator"
	"algo-trading/pkg/api/alphavantage"
)

// MovingAverageCrossover implements a simple moving average crossover strategy
type MovingAverageCrossover struct {
	avClient    *alphavantage.Client
	shortPeriod int
	longPeriod  int
}

// NewMovingAverageCrossover creates a new MA crossover strategy instance
func NewMovingAverageCrossover(client *alphavantage.Client, shortPeriod, longPeriod int) *MovingAverageCrossover {
	return &MovingAverageCrossover{
		avClient:    client,
		shortPeriod: shortPeriod,
		longPeriod:  longPeriod,
	}
}

// GetName returns the strategy name
func (ma *MovingAverageCrossover) GetName() string {
	return fmt.Sprintf("MA_CROSS_%d_%d", ma.shortPeriod, ma.longPeriod)
}

// GenerateSignal implements the TradingStrategy interface
func (ma *MovingAverageCrossover) GenerateSignal(quote *alphavantage.QuoteResponse) (*signalgenerator.TradingSignal, error) {
	// Get current price
	price, err := strconv.ParseFloat(quote.GlobalQuote.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse current price: %v", err)
	}

	// Get short-term MA
	shortMA, err := ma.avClient.GetTechnicalIndicator(
		quote.GlobalQuote.Symbol,
		"SMA",
		"daily",
		ma.shortPeriod,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get short-term MA: %v", err)
	}

	// Get long-term MA
	longMA, err := ma.avClient.GetTechnicalIndicator(
		quote.GlobalQuote.Symbol,
		"SMA",
		"daily",
		ma.longPeriod,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get long-term MA: %v", err)
	}

	// Get the latest values
	var shortValue, longValue float64
	for _, v := range shortMA.TechnicalIndicator {
		shortValue, _ = strconv.ParseFloat(v.Value, 64)
		break // Get first (latest) value
	}
	for _, v := range longMA.TechnicalIndicator {
		longValue, _ = strconv.ParseFloat(v.Value, 64)
		break // Get first (latest) value
	}

	// Generate signal based on MA crossover
	var signal *signalgenerator.TradingSignal

	if shortValue > longValue {
		// Short MA above long MA - potential buy signal
		signal = &signalgenerator.TradingSignal{
			Symbol:     quote.GlobalQuote.Symbol,
			Type:       signalgenerator.BuySignal,
			Price:      price,
			Timestamp:  time.Now(),
			Strategy:   ma.GetName(),
			Confidence: (shortValue - longValue) / longValue * 100,
		}
	} else if shortValue < longValue {
		// Short MA below long MA - potential sell signal
		signal = &signalgenerator.TradingSignal{
			Symbol:     quote.GlobalQuote.Symbol,
			Type:       signalgenerator.SellSignal,
			Price:      price,
			Timestamp:  time.Now(),
			Strategy:   ma.GetName(),
			Confidence: (longValue - shortValue) / longValue * 100,
		}
	}

	return signal, nil
}