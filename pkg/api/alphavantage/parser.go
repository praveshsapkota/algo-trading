package alphavantage

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseQuote parses a QuoteResponse into a ParsedQuote
func ParseQuote(resp *QuoteResponse) (*ParsedQuote, error) {
	if resp == nil {
		return nil, fmt.Errorf("nil response")
	}

	quote := resp.GlobalQuote
	if quote.Symbol == "" {
		return nil, fmt.Errorf("empty quote data")
	}

	open, err := strconv.ParseFloat(quote.Open, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing open: %w", err)
	}

	high, err := strconv.ParseFloat(quote.High, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing high: %w", err)
	}

	low, err := strconv.ParseFloat(quote.Low, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing low: %w", err)
	}

	price, err := strconv.ParseFloat(quote.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing price: %w", err)
	}

	volume, err := strconv.ParseInt(quote.Volume, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing volume: %w", err)
	}

	latestTradingDay, err := time.Parse("2006-01-02", quote.LatestTradingDay)
	if err != nil {
		return nil, fmt.Errorf("error parsing latest trading day: %w", err)
	}

	previousClose, err := strconv.ParseFloat(quote.PreviousClose, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing previous close: %w", err)
	}

	change, err := strconv.ParseFloat(quote.Change, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing change: %w", err)
	}

	// Remove the % sign and parse
	changePercentStr := strings.TrimSuffix(quote.ChangePercent, "%")
	changePercent, err := strconv.ParseFloat(changePercentStr, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing change percent: %w", err)
	}

	return &ParsedQuote{
		Symbol:           quote.Symbol,
		Open:             open,
		High:             high,
		Low:              low,
		Price:            price,
		Volume:           volume,
		LatestTradingDay: latestTradingDay,
		PreviousClose:    previousClose,
		Change:           change,
		ChangePercent:    changePercent,
	}, nil
}

// ParseDailyTimeSeries parses a TimeSeriesResponse into a slice of ParsedTimeSeriesItem
func ParseDailyTimeSeries(resp *TimeSeriesResponse) ([]ParsedTimeSeriesItem, error) {
	if resp == nil {
		return nil, fmt.Errorf("nil response")
	}

	if len(resp.TimeSeriesDaily) == 0 {
		return nil, fmt.Errorf("empty time series data")
	}

	result := make([]ParsedTimeSeriesItem, 0, len(resp.TimeSeriesDaily))

	for dateStr, item := range resp.TimeSeriesDaily {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing date %s: %w", dateStr, err)
		}

		open, err := strconv.ParseFloat(item.Open, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing open for %s: %w", dateStr, err)
		}

		high, err := strconv.ParseFloat(item.High, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing high for %s: %w", dateStr, err)
		}

		low, err := strconv.ParseFloat(item.Low, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing low for %s: %w", dateStr, err)
		}

		close, err := strconv.ParseFloat(item.Close, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing close for %s: %w", dateStr, err)
		}

		volume, err := strconv.ParseInt(item.Volume, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing volume for %s: %w", dateStr, err)
		}

		result = append(result, ParsedTimeSeriesItem{
			Date:   date,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: volume,
		})
	}

	// Sort by date in descending order (newest first)
	sortTimeSeriesByDate(result)

	return result, nil
}

// ParseIntraday parses an IntradayResponse into a slice of ParsedTimeSeriesItem
func ParseIntraday(resp *IntradayResponse) ([]ParsedTimeSeriesItem, error) {
	if resp == nil {
		return nil, fmt.Errorf("nil response")
	}

	if len(resp.TimeSeries) == 0 {
		return nil, fmt.Errorf("empty intraday data")
	}

	result := make([]ParsedTimeSeriesItem, 0, len(resp.TimeSeries))

	for dateTimeStr, item := range resp.TimeSeries {
		// Parse datetime in format "2023-04-13 09:30:00"
		dateTime, err := time.Parse("2006-01-02 15:04:05", dateTimeStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing datetime %s: %w", dateTimeStr, err)
		}

		open, err := strconv.ParseFloat(item.Open, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing open for %s: %w", dateTimeStr, err)
		}

		high, err := strconv.ParseFloat(item.High, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing high for %s: %w", dateTimeStr, err)
		}

		low, err := strconv.ParseFloat(item.Low, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing low for %s: %w", dateTimeStr, err)
		}

		close, err := strconv.ParseFloat(item.Close, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing close for %s: %w", dateTimeStr, err)
		}

		volume, err := strconv.ParseInt(item.Volume, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing volume for %s: %w", dateTimeStr, err)
		}

		result = append(result, ParsedTimeSeriesItem{
			Date:   dateTime,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: volume,
		})
	}

	// Sort by date in descending order (newest first)
	sortTimeSeriesByDate(result)

	return result, nil
}

// ParseTechnicalIndicator parses an IndicatorResponse into a slice of ParsedIndicatorItem
func ParseTechnicalIndicator(resp *IndicatorResponse) ([]ParsedIndicatorItem, error) {
	if resp == nil {
		return nil, fmt.Errorf("nil response")
	}

	if len(resp.TechnicalIndicator) == 0 {
		return nil, fmt.Errorf("empty indicator data")
	}

	result := make([]ParsedIndicatorItem, 0, len(resp.TechnicalIndicator))

	for dateStr, item := range resp.TechnicalIndicator {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			// Try intraday format
			date, err = time.Parse("2006-01-02 15:04:05", dateStr)
			if err != nil {
				return nil, fmt.Errorf("error parsing date %s: %w", dateStr, err)
			}
		}

		// Determine which field to use based on the indicator type
		var valueStr string
		if item.Value != "" {
			valueStr = item.Value
		} else if item.RSI != "" {
			valueStr = item.RSI
		} else if item.MACD != "" {
			valueStr = item.MACD
		} else {
			return nil, fmt.Errorf("no value found for indicator on %s", dateStr)
		}

		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing indicator value for %s: %w", dateStr, err)
		}

		result = append(result, ParsedIndicatorItem{
			Date:  date,
			Value: value,
		})
	}

	// Sort by date in descending order (newest first)
	sortIndicatorsByDate(result)

	return result, nil
}

// Helper function to sort time series items by date (newest first)
func sortTimeSeriesByDate(items []ParsedTimeSeriesItem) {
	// Sort by date in descending order (newest first)
	// Using a simple bubble sort for clarity, but in production code
	// you might want to use the sort package for better performance
	for i := 0; i < len(items)-1; i++ {
		for j := 0; j < len(items)-i-1; j++ {
			if items[j].Date.Before(items[j+1].Date) {
				items[j], items[j+1] = items[j+1], items[j]
			}
		}
	}
}

// Helper function to sort indicator items by date (newest first)
func sortIndicatorsByDate(items []ParsedIndicatorItem) {
	// Sort by date in descending order (newest first)
	for i := 0; i < len(items)-1; i++ {
		for j := 0; j < len(items)-i-1; j++ {
			if items[j].Date.Before(items[j+1].Date) {
				items[j], items[j+1] = items[j+1], items[j]
			}
		}
	}
}
