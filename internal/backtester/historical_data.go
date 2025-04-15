package backtester

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/user/algo-trading/pkg/api/alphavantage"
)

// HistoricalDataImportRequest represents a request to import historical data
type HistoricalDataImportRequest struct {
	Symbol    string `json:"symbol" binding:"required"`
	Timeframe string `json:"timeframe" binding:"required"`
	Source    string `json:"source" binding:"required"` // "alphavantage", "yahoo", "csv"
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	CSVData   string `json:"csv_data,omitempty"` // Base64 encoded CSV data if source is "csv"
}

// FetchHistoricalData fetches historical data from the specified source
func (bs *BacktesterService) FetchHistoricalData(ctx context.Context, request *HistoricalDataImportRequest) ([]HistoricalData, error) {
	switch request.Source {
	case "alphavantage":
		return bs.fetchFromAlphaVantage(ctx, request)
	case "yahoo":
		return bs.fetchFromYahooFinance(ctx, request)
	case "csv":
		return bs.parseCSVData(request)
	default:
		return nil, fmt.Errorf("unsupported data source: %s", request.Source)
	}
}

// fetchFromAlphaVantage fetches historical data from Alpha Vantage
func (bs *BacktesterService) fetchFromAlphaVantage(ctx context.Context, request *HistoricalDataImportRequest) ([]HistoricalData, error) {
	// Create Alpha Vantage client
	apiKey := bs.getAPIKey("ALPHA_VANTAGE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("Alpha Vantage API key not found")
	}

	client := alphavantage.NewClient(apiKey)

	// Map timeframe to Alpha Vantage interval
	interval := "daily"
	switch request.Timeframe {
	case "1m":
		interval = "1min"
	case "5m":
		interval = "5min"
	case "15m":
		interval = "15min"
	case "30m":
		interval = "30min"
	case "1h":
		interval = "60min"
	case "1d":
		interval = "daily"
	case "1w":
		interval = "weekly"
	case "1M":
		interval = "monthly"
	}

	// Fetch time series data
	var data []HistoricalData
	var err error

	if interval == "daily" || interval == "weekly" || interval == "monthly" {
		// Fetch daily, weekly, or monthly data
		timeSeries, err := client.GetTimeSeries(request.Symbol, interval, "full")
		if err != nil {
			return nil, fmt.Errorf("failed to fetch time series data: %v", err)
		}

		// Parse response
		for date, values := range timeSeries.TimeSeries {
			timestamp, err := time.Parse("2006-01-02", date)
			if err != nil {
				bs.logger.Warnf("Failed to parse date %s: %v", date, err)
				continue
			}

			// Filter by date range if specified
			if request.StartDate != "" {
				startDate, err := time.Parse("2006-01-02", request.StartDate)
				if err == nil && timestamp.Before(startDate) {
					continue
				}
			}
			if request.EndDate != "" {
				endDate, err := time.Parse("2006-01-02", request.EndDate)
				if err == nil && timestamp.After(endDate) {
					continue
				}
			}

			// Parse OHLCV values
			open, _ := strconv.ParseFloat(values.Open, 64)
			high, _ := strconv.ParseFloat(values.High, 64)
			low, _ := strconv.ParseFloat(values.Low, 64)
			close, _ := strconv.ParseFloat(values.Close, 64)
			volume, _ := strconv.ParseInt(values.Volume, 10, 64)

			data = append(data, HistoricalData{
				Symbol:    request.Symbol,
				Timestamp: timestamp,
				Open:      open,
				High:      high,
				Low:       low,
				Close:     close,
				Volume:    volume,
				Timeframe: request.Timeframe,
			})
		}
	} else {
		// Fetch intraday data
		timeSeries, err := client.GetIntradayTimeSeries(request.Symbol, interval, "full")
		if err != nil {
			return nil, fmt.Errorf("failed to fetch intraday time series data: %v", err)
		}

		// Parse response
		for datetime, values := range timeSeries.TimeSeries {
			timestamp, err := time.Parse("2006-01-02 15:04:05", datetime)
			if err != nil {
				bs.logger.Warnf("Failed to parse datetime %s: %v", datetime, err)
				continue
			}

			// Filter by date range if specified
			if request.StartDate != "" {
				startDate, err := time.Parse("2006-01-02", request.StartDate)
				if err == nil && timestamp.Before(startDate) {
					continue
				}
			}
			if request.EndDate != "" {
				endDate, err := time.Parse("2006-01-02", request.EndDate)
				if err == nil && timestamp.After(endDate) {
					continue
				}
			}

			// Parse OHLCV values
			open, _ := strconv.ParseFloat(values.Open, 64)
			high, _ := strconv.ParseFloat(values.High, 64)
			low, _ := strconv.ParseFloat(values.Low, 64)
			close, _ := strconv.ParseFloat(values.Close, 64)
			volume, _ := strconv.ParseInt(values.Volume, 10, 64)

			data = append(data, HistoricalData{
				Symbol:    request.Symbol,
				Timestamp: timestamp,
				Open:      open,
				High:      high,
				Low:       low,
				Close:     close,
				Volume:    volume,
				Timeframe: request.Timeframe,
			})
		}
	}

	return data, nil
}

// fetchFromYahooFinance fetches historical data from Yahoo Finance
func (bs *BacktesterService) fetchFromYahooFinance(ctx context.Context, request *HistoricalDataImportRequest) ([]HistoricalData, error) {
	// TODO: Implement Yahoo Finance API integration
	// This is a placeholder for future implementation
	return nil, fmt.Errorf("Yahoo Finance integration not implemented yet")
}

// parseCSVData parses CSV data for historical prices
func (bs *BacktesterService) parseCSVData(request *HistoricalDataImportRequest) ([]HistoricalData, error) {
	if request.CSVData == "" {
		return nil, fmt.Errorf("CSV data is empty")
	}

	// Create CSV reader
	reader := csv.NewReader(io.StringReader(request.CSVData))

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %v", err)
	}

	// Map header columns
	colMap := make(map[string]int)
	for i, col := range header {
		colMap[col] = i
	}

	// Check required columns
	requiredCols := []string{"timestamp", "open", "high", "low", "close", "volume"}
	for _, col := range requiredCols {
		if _, exists := colMap[col]; !exists {
			return nil, fmt.Errorf("required column '%s' not found in CSV", col)
		}
	}

	// Parse data rows
	var data []HistoricalData
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %v", err)
		}

		// Parse timestamp
		timestamp, err := time.Parse("2006-01-02 15:04:05", row[colMap["timestamp"]])
		if err != nil {
			// Try alternative date format
			timestamp, err = time.Parse("2006-01-02", row[colMap["timestamp"]])
			if err != nil {
				bs.logger.Warnf("Failed to parse timestamp %s: %v", row[colMap["timestamp"]], err)
				continue
			}
		}

		// Filter by date range if specified
		if request.StartDate != "" {
			startDate, err := time.Parse("2006-01-02", request.StartDate)
			if err == nil && timestamp.Before(startDate) {
				continue
			}
		}
		if request.EndDate != "" {
			endDate, err := time.Parse("2006-01-02", request.EndDate)
			if err == nil && timestamp.After(endDate) {
				continue
			}
		}

		// Parse OHLCV values
		open, _ := strconv.ParseFloat(row[colMap["open"]], 64)
		high, _ := strconv.ParseFloat(row[colMap["high"]], 64)
		low, _ := strconv.ParseFloat(row[colMap["low"]], 64)
		close, _ := strconv.ParseFloat(row[colMap["close"]], 64)
		volume, _ := strconv.ParseInt(row[colMap["volume"]], 10, 64)

		data = append(data, HistoricalData{
			Symbol:    request.Symbol,
			Timestamp: timestamp,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
			Timeframe: request.Timeframe,
		})
	}

	return data, nil
}

// StoreHistoricalData stores historical data in the database
func (bs *BacktesterService) StoreHistoricalData(ctx context.Context, data []HistoricalData) error {
	if len(data) == 0 {
		return nil
	}

	// Use batch insert with on conflict do nothing
	return bs.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "symbol"}, {Name: "timestamp"}, {Name: "timeframe"}},
		DoNothing: true,
	}).CreateInBatches(data, 1000).Error
}

// GetHistoricalData retrieves historical data from the database
func (bs *BacktesterService) GetHistoricalData(ctx context.Context, symbol, timeframe string, startDate, endDate time.Time) ([]HistoricalData, error) {
	var data []HistoricalData
	query := bs.db.WithContext(ctx).Where("symbol = ? AND timeframe = ?", symbol, timeframe)

	if !startDate.IsZero() {
		query = query.Where("timestamp >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("timestamp <= ?", endDate)
	}

	if err := query.Order("timestamp").Find(&data).Error; err != nil {
		return nil, err
	}

	return data, nil
}

// importHistoricalData handles the import historical data API endpoint
func (bs *BacktesterService) importHistoricalData(c *gin.Context) {
	var request HistoricalDataImportRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Fetch historical data
	data, err := bs.FetchHistoricalData(c.Request.Context(), &request)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to fetch historical data: %v", err)})
		return
	}

	// Store historical data
	if err := bs.StoreHistoricalData(c.Request.Context(), data); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to store historical data: %v", err)})
		return
	}

	c.JSON(200, gin.H{
		"message": fmt.Sprintf("Successfully imported %d data points for %s", len(data), request.Symbol),
		"count":   len(data),
	})
}

// getAPIKey retrieves an API key from environment variables
func (bs *BacktesterService) getAPIKey(key string) string {
	// In a real implementation, this would retrieve from a secure source
	// For now, we'll use a mock implementation
	if key == "ALPHA_VANTAGE_API_KEY" {
		return "demo" // Alpha Vantage demo key
	}
	return ""
}