package models

import (
	"time"
)

// OptionType represents the type of option
type OptionType string

const (
	Call OptionType = "CALL"
	Put  OptionType = "PUT"
)

// Option represents an option contract
type Option struct {
	ID               uint       `json:"id" gorm:"primaryKey"`
	UnderlyingSymbol string     `json:"underlying_symbol" gorm:"index"`
	Type             OptionType `json:"type"`
	StrikePrice      float64    `json:"strike_price"`
	ExpirationDate   time.Time  `json:"expiration_date" gorm:"index"`
	Bid              float64    `json:"bid"`
	Ask              float64    `json:"ask"`
	LastPrice        float64    `json:"last_price"`
	Volume           int        `json:"volume"`
	OpenInterest     int        `json:"open_interest"`
	ImpliedVolatility float64   `json:"implied_volatility"`
	Delta            float64    `json:"delta"`
	Gamma            float64    `json:"gamma"`
	Theta            float64    `json:"theta"`
	Vega             float64    `json:"vega"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// OptionsChain represents a collection of options for a specific underlying asset
type OptionsChain struct {
	UnderlyingSymbol string    `json:"underlying_symbol"`
	UnderlyingPrice  float64   `json:"underlying_price"`
	UpdatedAt        time.Time `json:"updated_at"`
	Calls            []Option  `json:"calls"`
	Puts             []Option  `json:"puts"`
}

// OptionPosition represents an open option position
type OptionPosition struct {
	ID               uint       `json:"id" gorm:"primaryKey"`
	UnderlyingSymbol string     `json:"underlying_symbol" gorm:"index"`
	Type             OptionType `json:"type"`
	StrikePrice      float64    `json:"strike_price"`
	ExpirationDate   time.Time  `json:"expiration_date"`
	Quantity         int        `json:"quantity"`
	AveragePrice     float64    `json:"average_price"`
	CurrentPrice     float64    `json:"current_price"`
	PnL              float64    `json:"pnl"`
	OpenDate         time.Time  `json:"open_date"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// OptionOrder represents an option order
type OptionOrder struct {
	ID               uint       `json:"id" gorm:"primaryKey"`
	UnderlyingSymbol string     `json:"underlying_symbol"`
	Type             OptionType `json:"type"`
	StrikePrice      float64    `json:"strike_price"`
	ExpirationDate   time.Time  `json:"expiration_date"`
	OrderType        string     `json:"order_type"` // BUY, SELL
	Quantity         int        `json:"quantity"`
	Price            float64    `json:"price"`
	Status           string     `json:"status"` // PENDING, EXECUTED, CANCELLED
	CreatedAt        time.Time  `json:"created_at"`
	ExecutedAt       *time.Time `json:"executed_at,omitempty"`
}

// OptionTrade represents a completed option trade
type OptionTrade struct {
	ID               uint       `json:"id" gorm:"primaryKey"`
	UnderlyingSymbol string     `json:"underlying_symbol"`
	Type             OptionType `json:"type"`
	StrikePrice      float64    `json:"strike_price"`
	ExpirationDate   time.Time  `json:"expiration_date"`
	TradeType        string     `json:"trade_type"` // BUY, SELL
	Quantity         int        `json:"quantity"`
	Price            float64    `json:"price"`
	Commission       float64    `json:"commission"`
	TradeDate        time.Time  `json:"trade_date"`
	Strategy         string     `json:"strategy,omitempty"`
}

// OptionStrategy represents an options trading strategy
type OptionStrategy struct {
	ID               uint       `json:"id" gorm:"primaryKey"`
	Name             string     `json:"name"`
	UnderlyingSymbol string     `json:"underlying_symbol"`
	Legs             []OptionLeg `json:"legs" gorm:"foreignKey:StrategyID"`
	MaxProfit        float64    `json:"max_profit"`
	MaxLoss          float64    `json:"max_loss"`
	BreakEven        float64    `json:"break_even"`
	CreatedAt        time.Time  `json:"created_at"`
}

// OptionLeg represents a leg in an options strategy
type OptionLeg struct {
	ID               uint       `json:"id" gorm:"primaryKey"`
	StrategyID       uint       `json:"strategy_id"`
	Type             OptionType `json:"type"`
	StrikePrice      float64    `json:"strike_price"`
	ExpirationDate   time.Time  `json:"expiration_date"`
	Action           string     `json:"action"` // BUY, SELL
	Quantity         int        `json:"quantity"`
}