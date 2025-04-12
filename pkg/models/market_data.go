package models

import (
	"time"
)

// StockQuote represents real-time stock quote data
type StockQuote struct {
	Symbol        string    `json:"symbol" gorm:"primaryKey;type:varchar(20)"`
	Price         float64   `json:"price" gorm:"type:decimal(10,2)"`
	Volume        int64     `json:"volume"`
	Change        float64   `json:"change" gorm:"type:decimal(10,2)"`
	ChangePercent float64   `json:"change_percent" gorm:"type:decimal(10,2)"`
	High          float64   `json:"high" gorm:"type:decimal(10,2)"`
	Low           float64   `json:"low" gorm:"type:decimal(10,2)"`
	Open          float64   `json:"open" gorm:"type:decimal(10,2)"`
	PreviousClose float64   `json:"previous_close" gorm:"type:decimal(10,2)"`
	Timestamp     time.Time `json:"timestamp" gorm:"index"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// HistoricalPrice represents historical price data for a stock
type HistoricalPrice struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Symbol    string    `json:"symbol" gorm:"type:varchar(20);index"`
	Date      time.Time `json:"date" gorm:"index"`
	Open      float64   `json:"open" gorm:"type:decimal(10,2)"`
	High      float64   `json:"high" gorm:"type:decimal(10,2)"`
	Low       float64   `json:"low" gorm:"type:decimal(10,2)"`
	Close     float64   `json:"close" gorm:"type:decimal(10,2)"`
	Volume    int64     `json:"volume"`
	Adjusted  float64   `json:"adjusted" gorm:"type:decimal(10,2)"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TechnicalIndicator represents calculated technical indicators for a stock
type TechnicalIndicator struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Symbol    string    `json:"symbol" gorm:"type:varchar(20);index"`
	Type      string    `json:"type" gorm:"type:varchar(50);index"` // e.g., "SMA", "EMA", "RSI"
	Period    int       `json:"period"`                             // e.g., 14 for 14-day RSI
	Value     float64   `json:"value" gorm:"type:decimal(10,4)"`
	Timestamp time.Time `json:"timestamp" gorm:"index"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// MarketNews represents news articles related to stocks or the market
type MarketNews struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Title       string    `json:"title" gorm:"type:varchar(255)"`
	URL         string    `json:"url" gorm:"type:varchar(255)"`
	Source      string    `json:"source" gorm:"type:varchar(100)"`
	Summary     string    `json:"summary" gorm:"type:text"`
	Symbols     string    `json:"symbols" gorm:"type:varchar(255)"` // Comma-separated list of related symbols
	PublishedAt time.Time `json:"published_at" gorm:"index"`
	Sentiment   float64   `json:"sentiment" gorm:"type:decimal(5,2)"` // -1.0 to 1.0 sentiment score
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// Watchlist represents a list of stocks being monitored
type Watchlist struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"type:varchar(100)"`
	Symbols   string    `json:"symbols" gorm:"type:text"` // Comma-separated list of symbols
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// DataSource represents a market data source configuration
type DataSource struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"type:varchar(100);uniqueIndex"`
	Type      string    `json:"type" gorm:"type:varchar(50)"` // e.g., "API", "Websocket"
	URL       string    `json:"url" gorm:"type:varchar(255)"`
	APIKey    string    `json:"api_key" gorm:"type:varchar(255)"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	Priority  int       `json:"priority" gorm:"default:1"` // Lower number = higher priority
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
