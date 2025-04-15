package tradeexecutor

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"algo-trading/internal/risk-manager"
	"algo-trading/internal/signal-generator"
	"algo-trading/pkg/models"
)

// TradeStatus represents the current status of a trade
type TradeStatus string

const (
	Pending   TradeStatus = "PENDING"
	Executed  TradeStatus = "EXECUTED"
	Cancelled TradeStatus = "CANCELLED"
	Completed TradeStatus = "COMPLETED"
)

// Trade represents a trade execution
type Trade struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	Symbol        string         `json:"symbol"`
	Type          string         `json:"type"`
	Quantity      float64        `json:"quantity"`
	Price         float64        `json:"price"`
	Status        TradeStatus    `json:"status"`
	StopLoss      float64        `json:"stop_loss"`
	TakeProfit    float64        `json:"take_profit"`
	Strategy      string         `json:"strategy"`
	SignalID      string         `json:"signal_id"`
	ExecutionTime time.Time      `json:"execution_time"`
	ClosingTime   *time.Time     `json:"closing_time,omitempty"`
	PnL           float64        `json:"pnl"`
	Notes         string         `json:"notes,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// TradeExecutor handles trade execution and management
type TradeExecutor struct {
	db          *gorm.DB
	redisCache  *redis.Client
	riskManager *riskmanager.RiskManager
	logger      *logrus.Logger
	mutex       sync.RWMutex
}

// NewTradeExecutor creates a new instance of TradeExecutor
func NewTradeExecutor(
	db *gorm.DB,
	redisAddr string,
	riskManager *riskmanager.RiskManager,
	logger *logrus.Logger,
) (*TradeExecutor, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &TradeExecutor{
		db:          db,
		redisCache:  redisClient,
		riskManager: riskManager,
		logger:      logger,
	}, nil
}

// ExecuteTrade processes a trading signal and executes a trade if approved
func (te *TradeExecutor) ExecuteTrade(ctx context.Context, signal *signalgenerator.TradingSignal) error {
	// Perform risk assessment
	riskResult, err := te.riskManager.AssessRisk(ctx, signal)
	if err != nil {
		return fmt.Errorf("risk assessment failed: %v", err)
	}

	if !riskResult.Approved {
		te.logger.Infof("Trade rejected: %s", riskResult.Reason)
		return nil
	}

	// Create trade record
	trade := &Trade{
		Symbol:        signal.Symbol,
		Type:          string(signal.Type),
		Quantity:      riskResult.MaxPositionSize,
		Price:         signal.Price,
		Status:        Pending,
		StopLoss:      riskResult.StopLoss,
		TakeProfit:    riskResult.TakeProfit,
		Strategy:      signal.Strategy,
		SignalID:      fmt.Sprintf("%s-%d", signal.Strategy, signal.Timestamp.Unix()),
		ExecutionTime: time.Now(),
	}

	// Begin transaction
	tx := te.db.Begin()

	// Save trade to database
	if err := tx.Create(trade).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create trade record: %v", err)
	}

	// Execute the trade (mock implementation)
	if err := te.executeMockTrade(trade); err != nil {
		tx.Rollback()
		return fmt.Errorf("trade execution failed: %v", err)
	}

	// Update trade status
	trade.Status = Executed
	if err := tx.Save(trade).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update trade status: %v", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Store trade in Redis for quick access
	if err := te.cacheTradeData(ctx, trade); err != nil {
		te.logger.Errorf("Failed to cache trade data: %v", err)
	}

	return nil
}

// executeMockTrade simulates trade execution (replace with actual broker integration)
func (te *TradeExecutor) executeMockTrade(trade *Trade) error {
	// Simulate network latency
	time.Sleep(time.Millisecond * 100)

	// In a real implementation, this would interact with a broker's API
	return nil
}

// cacheTradeData stores trade information in Redis
func (te *TradeExecutor) cacheTradeData(ctx context.Context, trade *Trade) error {
	// Store trade data
	tradeKey := fmt.Sprintf("trade:%d", trade.ID)
	tradeData, err := json.Marshal(trade)
	if err != nil {
		return err
	}

	pipe := te.redisCache.Pipeline()

	// Store trade data with 24-hour expiration
	pipe.Set(ctx, tradeKey, tradeData, 24*time.Hour)

	// Add to symbol's open positions if trade is executed
	if trade.Status == Executed {
		positionsKey := fmt.Sprintf("positions:%s", trade.Symbol)
		pipe.SAdd(ctx, positionsKey, trade.ID)
	}

	_, err = pipe.Exec(ctx)
	return err
}

// GetOpenTrades retrieves all open trades
func (te *TradeExecutor) GetOpenTrades(ctx context.Context) ([]*Trade, error) {
	var trades []*Trade

	if err := te.db.Where("status = ?", Executed).Find(&trades).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch open trades: %v", err)
	}

	return trades, nil
}

// UpdateTrade updates an existing trade's status and PnL
func (te *TradeExecutor) UpdateTrade(ctx context.Context, tradeID uint, currentPrice float64) error {
	var trade Trade

	if err := te.db.First(&trade, tradeID).Error; err != nil {
		return fmt.Errorf("failed to fetch trade: %v", err)
	}

	// Calculate PnL
	pnl := te.calculatePnL(&trade, currentPrice)
	trade.PnL = pnl

	// Check if stop loss or take profit has been hit
	if te.shouldCloseTrade(&trade, currentPrice) {
		now := time.Now()
		trade.Status = Completed
		trade.ClosingTime = &now
	}

	// Update trade in database
	if err := te.db.Save(&trade).Error; err != nil {
		return fmt.Errorf("failed to update trade: %v", err)
	}

	// Update cache
	return te.cacheTradeData(ctx, &trade)
}

// calculatePnL calculates the current profit/loss for a trade
func (te *TradeExecutor) calculatePnL(trade *Trade, currentPrice float64) float64 {
	switch trade.Type {
	case string(signalgenerator.BuySignal):
		return (currentPrice - trade.Price) * trade.Quantity
	case string(signalgenerator.SellSignal):
		return (trade.Price - currentPrice) * trade.Quantity
	default:
		return 0
	}
}

// shouldCloseTrade checks if a trade should be closed based on stop loss or take profit
func (te *TradeExecutor) shouldCloseTrade(trade *Trade, currentPrice float64) bool {
	if trade.Status != Executed {
		return false
	}

	switch trade.Type {
	case string(signalgenerator.BuySignal):
		return currentPrice <= trade.StopLoss || currentPrice >= trade.TakeProfit
	case string(signalgenerator.SellSignal):
		return currentPrice >= trade.StopLoss || currentPrice <= trade.TakeProfit
	default:
		return false
	}
}