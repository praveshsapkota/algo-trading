package signalgenerator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"

	"algo-trading/pkg/api/alphavantage"
)

// SignalType represents different types of trading signals
type SignalType string

const (
	BuySignal  SignalType = "BUY"
	SellSignal SignalType = "SELL"
	HoldSignal SignalType = "HOLD"
)

// TradingSignal represents a generated trading signal
type TradingSignal struct {
	Symbol    string                 `json:"symbol"`
	Type      SignalType             `json:"type"`
	Price     float64                `json:"price"`
	Timestamp time.Time              `json:"timestamp"`
	Strategy  string                 `json:"strategy"`
	Confidence float64               `json:"confidence"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// SignalGenerator handles the generation of trading signals
type SignalGenerator struct {
	avClient   *alphavantage.Client
	redisCache *redis.Client
	logger     *logrus.Logger
	mutex      sync.RWMutex
	strategies map[string]TradingStrategy
}

// TradingStrategy defines the interface for implementing trading strategies
type TradingStrategy interface {
	GenerateSignal(data *alphavantage.QuoteResponse) (*TradingSignal, error)
	GetName() string
}

// NewSignalGenerator creates a new instance of SignalGenerator
func NewSignalGenerator(avClient *alphavantage.Client, redisAddr string, logger *logrus.Logger) (*SignalGenerator, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &SignalGenerator{
		avClient:   avClient,
		redisCache: redisClient,
		logger:     logger,
		strategies: make(map[string]TradingStrategy),
	}, nil
}

// RegisterStrategy adds a new trading strategy to the generator
func (sg *SignalGenerator) RegisterStrategy(strategy TradingStrategy) {
	sg.mutex.Lock()
	defer sg.mutex.Unlock()
	sg.strategies[strategy.GetName()] = strategy
}

// GenerateSignals processes market data and generates trading signals
func (sg *SignalGenerator) GenerateSignals(ctx context.Context, symbol string) ([]*TradingSignal, error) {
	// Fetch latest quote data
	quote, err := sg.avClient.GetQuote(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch quote data: %v", err)
	}

	var signals []*TradingSignal

	// Generate signals from all registered strategies
	sg.mutex.RLock()
	for _, strategy := range sg.strategies {
		signal, err := strategy.GenerateSignal(quote)
		if err != nil {
			sg.logger.Errorf("Strategy %s failed: %v", strategy.GetName(), err)
			continue
		}
		if signal != nil {
			signals = append(signals, signal)
		}
	}
	sg.mutex.RUnlock()

	// Store signals in Redis for analysis
	if len(signals) > 0 {
		if err := sg.storeSignals(ctx, signals); err != nil {
			sg.logger.Errorf("Failed to store signals: %v", err)
		}
	}

	return signals, nil
}

// storeSignals stores generated signals in Redis
func (sg *SignalGenerator) storeSignals(ctx context.Context, signals []*TradingSignal) error {
	pipe := sg.redisCache.Pipeline()

	for _, signal := range signals {
		key := fmt.Sprintf("signal:%s:%s:%d",
			signal.Symbol,
			signal.Strategy,
			signal.Timestamp.Unix())

		// Store signal with 24-hour expiration
		pipe.Set(ctx, key, signal, 24*time.Hour)
	}

	_, err := pipe.Exec(ctx)
	return err
}