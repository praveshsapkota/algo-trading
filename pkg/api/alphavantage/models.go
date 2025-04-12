package alphavantage

import (
	"encoding/json"
	"time"
)

// BaseResponse contains common fields for all Alpha Vantage responses
type BaseResponse struct {
	ErrorMessage string `json:"Error Message,omitempty"`
	Information  string `json:"Information,omitempty"`
	Note         string `json:"Note,omitempty"`
}

// QuoteResponse represents the response from the GLOBAL_QUOTE endpoint
type QuoteResponse struct {
	BaseResponse
	GlobalQuote struct {
		Symbol           string `json:"01. symbol"`
		Open             string `json:"02. open"`
		High             string `json:"03. high"`
		Low              string `json:"04. low"`
		Price            string `json:"05. price"`
		Volume           string `json:"06. volume"`
		LatestTradingDay string `json:"07. latest trading day"`
		PreviousClose    string `json:"08. previous close"`
		Change           string `json:"09. change"`
		ChangePercent    string `json:"10. change percent"`
	} `json:"Global Quote"`
}

// TimeSeriesResponse represents the response from the TIME_SERIES_DAILY endpoint
type TimeSeriesResponse struct {
	BaseResponse
	MetaData struct {
		Information   string `json:"1. Information"`
		Symbol        string `json:"2. Symbol"`
		LastRefreshed string `json:"3. Last Refreshed"`
		OutputSize    string `json:"4. Output Size"`
		TimeZone      string `json:"5. Time Zone"`
	} `json:"Meta Data"`
	TimeSeriesDaily map[string]struct {
		Open   string `json:"1. open"`
		High   string `json:"2. high"`
		Low    string `json:"3. low"`
		Close  string `json:"4. close"`
		Volume string `json:"5. volume"`
	} `json:"Time Series (Daily)"`
}

// IntradayResponse represents the response from the TIME_SERIES_INTRADAY endpoint
type IntradayResponse struct {
	BaseResponse
	MetaData struct {
		Information   string `json:"1. Information"`
		Symbol        string `json:"2. Symbol"`
		LastRefreshed string `json:"3. Last Refreshed"`
		Interval      string `json:"4. Interval"`
		OutputSize    string `json:"5. Output Size"`
		TimeZone      string `json:"6. Time Zone"`
	} `json:"Meta Data"`
	TimeSeries map[string]struct {
		Open   string `json:"1. open"`
		High   string `json:"2. high"`
		Low    string `json:"3. low"`
		Close  string `json:"4. close"`
		Volume string `json:"5. volume"`
	} `json:"Time Series (1min)"` // This key changes based on the interval
}

// UnmarshalJSON custom unmarshaler for IntradayResponse to handle dynamic key
func (r *IntradayResponse) UnmarshalJSON(data []byte) error {
	// First, unmarshal the metadata
	var temp struct {
		BaseResponse
		MetaData struct {
			Information   string `json:"1. Information"`
			Symbol        string `json:"2. Symbol"`
			LastRefreshed string `json:"3. Last Refreshed"`
			Interval      string `json:"4. Interval"`
			OutputSize    string `json:"5. Output Size"`
			TimeZone      string `json:"6. Time Zone"`
		} `json:"Meta Data"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	r.BaseResponse = temp.BaseResponse
	r.MetaData = temp.MetaData

	// Now, determine the time series key based on the interval
	var timeSeriesKey string
	switch r.MetaData.Interval {
	case "1min":
		timeSeriesKey = "Time Series (1min)"
	case "5min":
		timeSeriesKey = "Time Series (5min)"
	case "15min":
		timeSeriesKey = "Time Series (15min)"
	case "30min":
		timeSeriesKey = "Time Series (30min)"
	case "60min":
		timeSeriesKey = "Time Series (60min)"
	default:
		timeSeriesKey = "Time Series (1min)" // Default
	}

	// Unmarshal the time series data with the correct key
	var timeSeriesData map[string]map[string]struct {
		Open   string `json:"1. open"`
		High   string `json:"2. high"`
		Low    string `json:"3. low"`
		Close  string `json:"4. close"`
		Volume string `json:"5. volume"`
	}

	if err := json.Unmarshal(data, &timeSeriesData); err != nil {
		return err
	}

	r.TimeSeries = timeSeriesData[timeSeriesKey]
	return nil
}

// IndicatorResponse represents the response from technical indicator endpoints
type IndicatorResponse struct {
	BaseResponse
	MetaData struct {
		Symbol        string `json:"1: Symbol"`
		Indicator     string `json:"2: Indicator"`
		LastRefreshed string `json:"3: Last Refreshed"`
		Interval      string `json:"4: Interval"`
		TimePeriod    string `json:"5: Time Period"`
		SeriesType    string `json:"6: Series Type"`
		TimeZone      string `json:"7: Time Zone"`
	} `json:"Meta Data"`
	TechnicalIndicator map[string]struct {
		Value string `json:"value,omitempty"` // For SMA, EMA, etc.
		RSI   string `json:"RSI,omitempty"`   // For RSI
		MACD  string `json:"MACD,omitempty"`  // For MACD
		// Add other indicator-specific fields as needed
	} `json:"Technical Analysis"`
}

// UnmarshalJSON custom unmarshaler for IndicatorResponse to handle dynamic keys
func (r *IndicatorResponse) UnmarshalJSON(data []byte) error {
	// First, unmarshal the metadata
	var temp struct {
		BaseResponse
		MetaData struct {
			Symbol        string `json:"1: Symbol"`
			Indicator     string `json:"2: Indicator"`
			LastRefreshed string `json:"3: Last Refreshed"`
			Interval      string `json:"4: Interval"`
			TimePeriod    string `json:"5: Time Period"`
			SeriesType    string `json:"6: Series Type"`
			TimeZone      string `json:"7: Time Zone"`
		} `json:"Meta Data"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	r.BaseResponse = temp.BaseResponse
	r.MetaData = temp.MetaData

	// Now, determine the technical indicator key based on the indicator
	var indicatorKey string
	switch r.MetaData.Indicator {
	case "Simple Moving Average (SMA)":
		indicatorKey = "Technical Analysis: SMA"
	case "Exponential Moving Average (EMA)":
		indicatorKey = "Technical Analysis: EMA"
	case "Relative Strength Index (RSI)":
		indicatorKey = "Technical Analysis: RSI"
	case "Moving Average Convergence/Divergence (MACD)":
		indicatorKey = "Technical Analysis: MACD"
	default:
		indicatorKey = "Technical Analysis: SMA" // Default
	}

	// Unmarshal the technical indicator data with the correct key
	var indicatorData map[string]map[string]struct {
		Value string `json:"value,omitempty"` // For SMA, EMA, etc.
		RSI   string `json:"RSI,omitempty"`   // For RSI
		MACD  string `json:"MACD,omitempty"`  // For MACD
		// Add other indicator-specific fields as needed
	}

	if err := json.Unmarshal(data, &indicatorData); err != nil {
		return err
	}

	r.TechnicalIndicator = indicatorData[indicatorKey]
	return nil
}

// SearchResponse represents the response from the SYMBOL_SEARCH endpoint
type SearchResponse struct {
	BaseResponse
	BestMatches []struct {
		Symbol      string `json:"1. symbol"`
		Name        string `json:"2. name"`
		Type        string `json:"3. type"`
		Region      string `json:"4. region"`
		MarketOpen  string `json:"5. marketOpen"`
		MarketClose string `json:"6. marketClose"`
		Timezone    string `json:"7. timezone"`
		Currency    string `json:"8. currency"`
		MatchScore  string `json:"9. matchScore"`
	} `json:"bestMatches"`
}

// ParsedQuote represents a parsed stock quote with proper types
type ParsedQuote struct {
	Symbol           string
	Open             float64
	High             float64
	Low              float64
	Price            float64
	Volume           int64
	LatestTradingDay time.Time
	PreviousClose    float64
	Change           float64
	ChangePercent    float64
}

// ParsedTimeSeriesItem represents a parsed time series item with proper types
type ParsedTimeSeriesItem struct {
	Date   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
}

// ParsedIndicatorItem represents a parsed technical indicator item with proper types
type ParsedIndicatorItem struct {
	Date  time.Time
	Value float64
}
