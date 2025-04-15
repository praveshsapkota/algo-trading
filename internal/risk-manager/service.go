package riskmanager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"

	"algo-trading/internal/signal-generator"
)

// RiskCheckResult represents the outcome of a risk assessment
type RiskCheckResult struct {
	Approved     bool    `json:"approved"`
	Reason       string  `json:"reason,omitempty"`
	RiskScore    float64 `json:"risk_score"`
	MaxPositionSize float64 `json:"max_position_size"`
	StopLoss     float64 `json:"stop_loss,omitempty"`
	TakeProfit   float64 `json:"take_profit,omitempty"`
}

// RiskConfig holds risk management parameters
type RiskConfig struct {
	MaxPositionSize     float64 `json:"max_position_size"`
	MaxDrawdown         float64 `json:"max_drawdown"`
	RiskPerTrade        float64 `json:"risk_per_trade"`
	MaxDailyLoss        float64 `json:"max_daily_loss"`
	MaxOpenPositions    int     `json:"max_open_positions"`
	MinConfidenceScore  float64 `json:"min_confidence_score"`
}

// RiskManager handles risk assessment and management
type RiskManager struct {
	config     RiskConfig
	redisCache *redis.Client
	logger     *logrus.Logger
	mutex      sync.RWMutex
}

// NewRiskManager creates a new instance of RiskManager
func NewRiskManager(config RiskConfig, redisAddr string, logger *logrus.Logger) (*RiskManager, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &RiskManager{
		config:     config,
		redisCache: redisClient,
		logger:     logger,
	}, nil
}

// AssessRisk evaluates a trading signal against risk parameters
func (rm *RiskManager) AssessRisk(ctx context.Context, signal *signalgenerator.TradingSignal) (*RiskCheckResult, error) {
	// Check signal confidence
	if signal.Confidence < rm.config.MinConfidenceScore {
		return &RiskCheckResult{
			Approved:  false,
			Reason:    "Signal confidence below threshold",
			RiskScore: signal.Confidence,
		}, nil
	}

	// Check current open positions
	openPositions, err := rm.getOpenPositionsCount(ctx, signal.Symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to check open positions: %v", err)
	}

	if openPositions >= rm.config.MaxOpenPositions {
		return &RiskCheckResult{
			Approved:  false,
			Reason:    "Maximum open positions reached",
			RiskScore: 100,
		}, nil
	}

	// Check daily loss limit
	dailyLoss, err := rm.getDailyLoss(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check daily loss: %v", err)
	}

	if dailyLoss >= rm.config.MaxDailyLoss {
		return &RiskCheckResult{
			Approved:  false,
			Reason:    "Daily loss limit reached",
			RiskScore: 100,
		}, nil
	}

	// Calculate position size based on risk per trade
	maxPositionSize := rm.calculatePositionSize(signal.Price)

	// Calculate stop loss and take profit levels
	stopLoss, takeProfit := rm.calculateExitPoints(signal)

	return &RiskCheckResult{
		Approved:        true,
		RiskScore:       rm.calculateRiskScore(signal),
		MaxPositionSize: maxPositionSize,
		StopLoss:        stopLoss,
		TakeProfit:      takeProfit,
	}, nil
}

// getOpenPositionsCount returns the number of open positions for a symbol
func (rm *RiskManager) getOpenPositionsCount(ctx context.Context, symbol string) (int, error) {
	key := fmt.Sprintf("positions:%s", symbol)
	count, err := rm.redisCache.SCard(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// getDailyLoss returns the current daily loss
func (rm *RiskManager) getDailyLoss(ctx context.Context) (float64, error) {
	key := fmt.Sprintf("daily_loss:%s", time.Now().Format("2006-01-02"))
	loss, err := rm.redisCache.Get(ctx, key).Float64()
	if err == redis.Nil {
		return 0, nil
	}
	return loss, err
}

// calculatePositionSize determines the appropriate position size based on risk parameters
func (rm *RiskManager) calculatePositionSize(price float64) float64 {
	// Calculate position size based on risk per trade and current price
	positionSize := (rm.config.RiskPerTrade * rm.config.MaxPositionSize) / price

	// Ensure position size doesn't exceed maximum
	if positionSize > rm.config.MaxPositionSize {
		positionSize = rm.config.MaxPositionSize
	}

	return positionSize
}

// calculateExitPoints determines stop loss and take profit levels
func (rm *RiskManager) calculateExitPoints(signal *signalgenerator.TradingSignal) (stopLoss, takeProfit float64) {
	// Calculate stop loss and take profit based on signal type and confidence
	volatilityFactor := 0.02 // 2% default volatility factor

	switch signal.Type {
	case signalgenerator.BuySignal:
		stopLoss = signal.Price * (1 - volatilityFactor)
		takeProfit = signal.Price * (1 + volatilityFactor*2) // 1:2 risk/reward ratio
	case signalgenerator.SellSignal:
		stopLoss = signal.Price * (1 + volatilityFactor)
		takeProfit = signal.Price * (1 - volatilityFactor*2) // 1:2 risk/reward ratio
	}

	return stopLoss, takeProfit
}

// calculateRiskScore computes a risk score for the trade
func (rm *RiskManager) calculateRiskScore(signal *signalgenerator.TradingSignal) float64 {
	// Basic risk score based on signal confidence and market conditions
	baseScore := signal.Confidence

	// Adjust score based on time of day, market volatility, etc.
	// For now, using a simple calculation
	return baseScore
}