package backtester

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/user/algo-trading/internal/signal-generator"
)

// BacktesterService handles backtesting of trading strategies
type BacktesterService struct {
	db         *gorm.DB
	redisCache *redis.Client
	logger     *logrus.Logger
	router     *gin.Engine
	strategies map[string]signalgenerator.TradingStrategy
}

// BacktestRequest represents a request to run a backtest
type BacktestRequest struct {
	StrategyName string                 `json:"strategy_name" binding:"required"`
	Symbols      []string               `json:"symbols" binding:"required"`
	StartDate    time.Time              `json:"start_date" binding:"required"`
	EndDate      time.Time              `json:"end_date" binding:"required"`
	Parameters   map[string]interface{} `json:"parameters"`
	Timeframe    string                 `json:"timeframe" binding:"required"`
}

// BacktestResult represents the result of a backtest
type BacktestResult struct {
	ID           uint                   `json:"id" gorm:"primaryKey"`
	StrategyName string                 `json:"strategy_name"`
	Symbols      []string               `json:"symbols" gorm:"type:text[]"`
	StartDate    time.Time              `json:"start_date"`
	EndDate      time.Time              `json:"end_date"`
	Parameters   map[string]interface{} `json:"parameters" gorm:"type:jsonb"`
	TotalReturn  float64                `json:"total_return"`
	WinRate      float64                `json:"win_rate"`
	MaxDrawdown  float64                `json:"max_drawdown"`
	SharpeRatio  float64                `json:"sharpe_ratio"`
	TradeCount   int                    `json:"trade_count"`
	Trades       []BacktestTrade        `json:"trades,omitempty" gorm:"foreignKey:BacktestResultID"`
	CreatedAt    time.Time              `json:"created_at"`
}

// BacktestTrade represents a simulated trade in a backtest
type BacktestTrade struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	BacktestResultID uint      `json:"backtest_result_id"`
	Symbol           string    `json:"symbol"`
	Type             string    `json:"type"`
	EntryPrice       float64   `json:"entry_price"`
	ExitPrice        float64   `json:"exit_price"`
	Quantity         float64   `json:"quantity"`
	EntryTime        time.Time `json:"entry_time"`
	ExitTime         time.Time `json:"exit_time"`
	PnL              float64   `json:"pnl"`
	Strategy         string    `json:"strategy"`
}

// HistoricalData represents historical price data
type HistoricalData struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Symbol    string    `json:"symbol" gorm:"index:idx_symbol_timestamp_timeframe,unique"`
	Timestamp time.Time `json:"timestamp" gorm:"index:idx_symbol_timestamp_timeframe,unique"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    int64     `json:"volume"`
	Timeframe string    `json:"timeframe" gorm:"index:idx_symbol_timestamp_timeframe,unique"`
}

// NewBacktesterService creates a new instance of BacktesterService
func NewBacktesterService(
	dbHost, dbPort, dbUser, dbPassword, dbName, redisAddr string,
	logger *logrus.Logger,
) (*BacktesterService, error) {
	// Connect to database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Auto-migrate schemas
	if err := db.AutoMigrate(&BacktestResult{}, &BacktestTrade{}, &HistoricalData{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database schemas: %v", err)
	}

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	router := gin.Default()

	return &BacktesterService{
		db:         db,
		redisCache: redisClient,
		logger:     logger,
		router:     router,
		strategies: make(map[string]signalgenerator.TradingStrategy),
	}, nil
}

// SetupRoutes configures the API endpoints
func (bs *BacktesterService) SetupRoutes() {
	// API routes
	api := bs.router.Group("/api/backtest")
	{
		api.GET("/strategies", bs.getStrategies)
		api.POST("/run", bs.runBacktest)
		api.GET("/results/:id", bs.getBacktestResult)
		api.GET("/results", bs.getBacktestResults)
		api.GET("/data/:symbol", bs.getHistoricalData)
		api.POST("/data/import", bs.importHistoricalData)
	}
}

// Start starts the HTTP server
func (bs *BacktesterService) Start(addr string) error {
	return bs.router.Run(addr)
}

// RegisterStrategy adds a trading strategy to the backtester
func (bs *BacktesterService) RegisterStrategy(strategy signalgenerator.TradingStrategy) {
	bs.strategies[strategy.GetName()] = strategy
}

// getStrategies handles the strategies API endpoint
func (bs *BacktesterService) getStrategies(c *gin.Context) {
	var strategies []string
	for name := range bs.strategies {
		strategies = append(strategies, name)
	}
	c.JSON(200, strategies)
}

// runBacktest handles the run backtest API endpoint
func (bs *BacktesterService) runBacktest(c *gin.Context) {
	var request BacktestRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Check if strategy exists
	strategy, exists := bs.strategies[request.StrategyName]
	if !exists {
		c.JSON(400, gin.H{"error": "Strategy not found"})
		return
	}

	// Run backtest
	result, err := bs.RunBacktest(c.Request.Context(), &request, strategy)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, result)
}

// getBacktestResult handles the get backtest result API endpoint
func (bs *BacktesterService) getBacktestResult(c *gin.Context) {
	id := c.Param("id")
	var result BacktestResult
	if err := bs.db.Preload("Trades").First(&result, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Backtest result not found"})
		return
	}
	c.JSON(200, result)
}

// getBacktestResults handles the get backtest results API endpoint
func (bs *BacktesterService) getBacktestResults(c *gin.Context) {
	var results []BacktestResult
	if err := bs.db.Find(&results).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, results)
}

// getHistoricalData handles the get historical data API endpoint
func (bs *BacktesterService) getHistoricalData(c *gin.Context) {
	symbol := c.Param("symbol")
	timeframe := c.DefaultQuery("timeframe", "1d")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var data []HistoricalData
	query := bs.db.Where("symbol = ? AND timeframe = ?", symbol, timeframe)

	if startDate != "" {
		query = query.Where("timestamp >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("timestamp <= ?", endDate)
	}

	if err := query.Order("timestamp").Find(&data).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, data)
}

// importHistoricalData handles the import historical data API endpoint
func (bs *BacktesterService) importHistoricalData(c *gin.Context) {
	// Implementation will be added in historical_data.go
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

// RunBacktest executes a backtest for a given strategy and parameters
func (bs *BacktesterService) RunBacktest(ctx context.Context, request *BacktestRequest, strategy signalgenerator.TradingStrategy) (*BacktestResult, error) {
	// Implementation will be added in strategy_runner.go
	return nil, fmt.Errorf("not implemented yet")
}