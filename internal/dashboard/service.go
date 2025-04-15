package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"algo-trading/internal/signal-generator"
	"algo-trading/internal/trade-executor"
)

// DashboardService handles the web interface and data aggregation
type DashboardService struct {
	db         *gorm.DB
	redisCache *redis.Client
	logger     *logrus.Logger
	router     *gin.Engine
}

// DashboardStats represents aggregated trading statistics
type DashboardStats struct {
	TotalTrades      int     `json:"total_trades"`
	WinningTrades    int     `json:"winning_trades"`
	LosingTrades     int     `json:"losing_trades"`
	TotalPnL         float64 `json:"total_pnl"`
	WinRate          float64 `json:"win_rate"`
	AveragePnL       float64 `json:"average_pnl"`
	OpenPositions    int     `json:"open_positions"`
	DailyPnL         float64 `json:"daily_pnl"`
	FalseSignals     int     `json:"false_signals"`
	SignalAccuracy   float64 `json:"signal_accuracy"`
}

// NewDashboardService creates a new instance of DashboardService
func NewDashboardService(
	db *gorm.DB,
	redisAddr string,
	logger *logrus.Logger,
) (*DashboardService, error) {
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

	return &DashboardService{
		db:         db,
		redisCache: redisClient,
		logger:     logger,
		router:     router,
	}, nil
}

// SetupRoutes configures the dashboard API endpoints
func (ds *DashboardService) SetupRoutes() {
	// API routes
	api := ds.router.Group("/api")
	{
		// Dashboard statistics
		api.GET("/stats", ds.getStats)
		api.GET("/trades", ds.getTrades)
		api.GET("/signals", ds.getSignals)
		api.GET("/performance", ds.getPerformance)
		api.GET("/positions", ds.getOpenPositions)
		
		// Historical data
		api.GET("/history/trades", ds.getTradeHistory)
		api.GET("/history/signals", ds.getSignalHistory)
		api.GET("/history/pnl", ds.getPnLHistory)
	}

	// Serve static files
	ds.router.Static("/static", "./web/build/static")
	ds.router.StaticFile("/", "./web/build/index.html")
}

// Start starts the dashboard server
func (ds *DashboardService) Start(port string) error {
	return ds.router.Run(port)
}

// getStats handles the stats API endpoint
func (ds *DashboardService) getStats(c *gin.Context) {
	stats, err := ds.calculateStats(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, stats)
}

// calculateStats computes current trading statistics
func (ds *DashboardService) calculateStats(ctx context.Context) (*DashboardStats, error) {
	var stats DashboardStats

	// Get total trades
	if err := ds.db.Model(&tradeexecutor.Trade{}).Count(&stats.TotalTrades).Error; err != nil {
		return nil, err
	}

	// Get winning trades
	if err := ds.db.Model(&tradeexecutor.Trade{}).Where("pnl > 0").Count(&stats.WinningTrades).Error; err != nil {
		return nil, err
	}

	// Get losing trades
	if err := ds.db.Model(&tradeexecutor.Trade{}).Where("pnl < 0").Count(&stats.LosingTrades).Error; err != nil {
		return nil, err
	}

	// Calculate total PnL
	if err := ds.db.Model(&tradeexecutor.Trade{}).Select("COALESCE(SUM(pnl), 0)").Scan(&stats.TotalPnL).Error; err != nil {
		return nil, err
	}

	// Calculate win rate
	if stats.TotalTrades > 0 {
		stats.WinRate = float64(stats.WinningTrades) / float64(stats.TotalTrades) * 100
	}

	// Calculate average PnL
	if stats.TotalTrades > 0 {
		stats.AveragePnL = stats.TotalPnL / float64(stats.TotalTrades)
	}

	// Get open positions
	if err := ds.db.Model(&tradeexecutor.Trade{}).Where("status = ?", tradeexecutor.Executed).Count(&stats.OpenPositions).Error; err != nil {
		return nil, err
	}

	// Calculate daily PnL
	today := time.Now().Format("2006-01-02")
	if err := ds.db.Model(&tradeexecutor.Trade{}).Where("DATE(created_at) = ?", today).Select("COALESCE(SUM(pnl), 0)").Scan(&stats.DailyPnL).Error; err != nil {
		return nil, err
	}

	// Get false signals count from Redis
	falseSignalsKey := fmt.Sprintf("stats:false_signals:%s", today)
	falseSignals, err := ds.redisCache.Get(ctx, falseSignalsKey).Int()
	if err != redis.Nil && err != nil {
		return nil, err
	}
	stats.FalseSignals = falseSignals

	// Calculate signal accuracy
	totalSignalsKey := fmt.Sprintf("stats:total_signals:%s", today)
	totalSignals, err := ds.redisCache.Get(ctx, totalSignalsKey).Int()
	if err != redis.Nil && err != nil {
		return nil, err
	}

	if totalSignals > 0 {
		stats.SignalAccuracy = (1 - float64(falseSignals)/float64(totalSignals)) * 100
	}

	return &stats, nil
}

// getTrades handles the trades API endpoint
func (ds *DashboardService) getTrades(c *gin.Context) {
	var trades []tradeexecutor.Trade

	if err := ds.db.Order("created_at desc").Limit(100).Find(&trades).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, trades)
}

// getSignals handles the signals API endpoint
func (ds *DashboardService) getSignals(c *gin.Context) {
	var signals []signalgenerator.TradingSignal
	ctx := c.Request.Context()

	// Get recent signals from Redis
	keys, err := ds.redisCache.Keys(ctx, "signal:*").Result()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	for _, key := range keys {
		var signal signalgenerator.TradingSignal
		data, err := ds.redisCache.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}

		if err := json.Unmarshal(data, &signal); err != nil {
			continue
		}

		signals = append(signals, signal)
	}

	c.JSON(200, signals)
}

// getPerformance handles the performance API endpoint
func (ds *DashboardService) getPerformance(c *gin.Context) {
	// Get performance metrics for different timeframes
	daily, err := ds.getPerformanceMetrics("daily")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	weekly, err := ds.getPerformanceMetrics("weekly")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	monthly, err := ds.getPerformanceMetrics("monthly")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"daily":   daily,
		"weekly":  weekly,
		"monthly": monthly,
	})
}

// getPerformanceMetrics calculates performance metrics for a given timeframe
func (ds *DashboardService) getPerformanceMetrics(timeframe string) (map[string]interface{}, error) {
	// Implementation depends on specific metrics needed
	return map[string]interface{}{
		"pnl":        0,
		"trade_count": 0,
		"win_rate":    0,
	}, nil
}

// getOpenPositions handles the positions API endpoint
func (ds *DashboardService) getOpenPositions(c *gin.Context) {
	var positions []tradeexecutor.Trade

	if err := ds.db.Where("status = ?", tradeexecutor.Executed).Find(&positions).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, positions)
}

// getTradeHistory handles the trade history API endpoint
func (ds *DashboardService) getTradeHistory(c *gin.Context) {
	var trades []tradeexecutor.Trade

	if err := ds.db.Order("created_at desc").Find(&trades).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, trades)
}

// getSignalHistory handles the signal history API endpoint
func (ds *DashboardService) getSignalHistory(c *gin.Context) {
	// Implementation depends on how signals are stored
	c.JSON(200, []interface{}{})
}

// getPnLHistory handles the PnL history API endpoint
func (ds *DashboardService) getPnLHistory(c *gin.Context) {
	// Implementation depends on how PnL data is stored
	c.JSON(200, []interface{}{})
}