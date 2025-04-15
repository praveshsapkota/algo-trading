package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// MarketDataService handles fetching and processing market data
type MarketDataService struct {
	apiKey     string
	redisCache *redis.Client
	mutex      sync.RWMutex
	symbols    map[string]bool
}

// MarketData represents the structure of market data
type MarketData struct {
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Volume    int64     `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}

// NewMarketDataService creates a new instance of MarketDataService
func NewMarketDataService(apiKey string, redisAddr string) (*MarketDataService, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &MarketDataService{
		apiKey:     apiKey,
		redisCache: redisClient,
		symbols:    make(map[string]bool),
	}, nil
}

// Subscribe adds a symbol to the watchlist
func (s *MarketDataService) Subscribe(symbol string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.symbols[symbol] = true
}

// Unsubscribe removes a symbol from the watchlist
func (s *MarketDataService) Unsubscribe(symbol string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.symbols, symbol)
}

// ProcessMarketData processes incoming market data and stores it in Redis
func (s *MarketDataService) ProcessMarketData(data *MarketData) error {
	ctx := context.Background()

	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal market data: %v", err)
	}

	// Store in Redis with expiration
	key := fmt.Sprintf("market_data:%s", data.Symbol)
	if err := s.redisCache.Set(ctx, key, jsonData, time.Hour*24).Err(); err != nil {
		return fmt.Errorf("failed to store market data in Redis: %v", err)
	}

	// Publish to Redis channel for real-time updates
	channel := fmt.Sprintf("market_updates:%s", data.Symbol)
	if err := s.redisCache.Publish(ctx, channel, jsonData).Err(); err != nil {
		return fmt.Errorf("failed to publish market data: %v", err)
	}

	return nil
}

// StartDataFetching begins the market data fetching loop
func (s *MarketDataService) StartDataFetching(ctx context.Context) {
	ticker := time.NewTicker(time.Second) // Adjust frequency as needed
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.mutex.RLock()
			for symbol := range s.symbols {
				go func(sym string) {
					if err := s.fetchMarketData(sym); err != nil {
						log.Printf("Error fetching market data for %s: %v", sym, err)
					}
				}(symbol)
			}
			s.mutex.RUnlock()
		}
	}
}

// fetchMarketData fetches market data for a single symbol
func (s *MarketDataService) fetchMarketData(symbol string) error {
	// TODO: Implement Alpha Vantage API call
	// For now, return mock data
	data := &MarketData{
		Symbol:    symbol,
		Price:     100.0, // Mock price
		Volume:    1000,  // Mock volume
		Timestamp: time.Now(),
	}

	return s.ProcessMarketData(data)
}