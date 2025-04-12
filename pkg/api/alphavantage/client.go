package alphavantage

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	baseURL = "https://www.alphavantage.co/query"
)

// Client represents an Alpha Vantage API client
type Client struct {
	apiKey     string
	httpClient *http.Client
	logger     *logrus.Logger
}

// NewClient creates a new Alpha Vantage API client
func NewClient(apiKey string, logger *logrus.Logger) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// GetQuote fetches the latest quote for a symbol
func (c *Client) GetQuote(symbol string) (*QuoteResponse, error) {
	params := url.Values{}
	params.Add("function", "GLOBAL_QUOTE")
	params.Add("symbol", symbol)
	params.Add("apikey", c.apiKey)

	resp, err := c.makeRequest(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var quoteResp QuoteResponse
	if err := json.Unmarshal(body, &quoteResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling quote response: %w", err)
	}

	// Check for API error
	if quoteResp.ErrorMessage != "" {
		return nil, fmt.Errorf("alpha vantage API error: %s", quoteResp.ErrorMessage)
	}

	return &quoteResp, nil
}

// GetDailyTimeSeries fetches daily time series data for a symbol
func (c *Client) GetDailyTimeSeries(symbol string, outputSize string) (*TimeSeriesResponse, error) {
	params := url.Values{}
	params.Add("function", "TIME_SERIES_DAILY")
	params.Add("symbol", symbol)
	params.Add("outputsize", outputSize) // "compact" or "full"
	params.Add("apikey", c.apiKey)

	resp, err := c.makeRequest(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var tsResp TimeSeriesResponse
	if err := json.Unmarshal(body, &tsResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling time series response: %w", err)
	}

	// Check for API error
	if tsResp.ErrorMessage != "" {
		return nil, fmt.Errorf("alpha vantage API error: %s", tsResp.ErrorMessage)
	}

	return &tsResp, nil
}

// GetIntraday fetches intraday time series data for a symbol
func (c *Client) GetIntraday(symbol string, interval string, outputSize string) (*IntradayResponse, error) {
	params := url.Values{}
	params.Add("function", "TIME_SERIES_INTRADAY")
	params.Add("symbol", symbol)
	params.Add("interval", interval) // "1min", "5min", "15min", "30min", "60min"
	params.Add("outputsize", outputSize)
	params.Add("apikey", c.apiKey)

	resp, err := c.makeRequest(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var intrResp IntradayResponse
	if err := json.Unmarshal(body, &intrResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling intraday response: %w", err)
	}

	// Check for API error
	if intrResp.ErrorMessage != "" {
		return nil, fmt.Errorf("alpha vantage API error: %s", intrResp.ErrorMessage)
	}

	return &intrResp, nil
}

// GetTechnicalIndicator fetches technical indicator data for a symbol
func (c *Client) GetTechnicalIndicator(symbol, function, interval string, timePeriod int) (*IndicatorResponse, error) {
	params := url.Values{}
	params.Add("function", function) // e.g., "SMA", "EMA", "RSI"
	params.Add("symbol", symbol)
	params.Add("interval", interval)
	params.Add("time_period", fmt.Sprintf("%d", timePeriod))
	params.Add("series_type", "close")
	params.Add("apikey", c.apiKey)

	resp, err := c.makeRequest(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var indResp IndicatorResponse
	if err := json.Unmarshal(body, &indResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling indicator response: %w", err)
	}

	// Check for API error
	if indResp.ErrorMessage != "" {
		return nil, fmt.Errorf("alpha vantage API error: %s", indResp.ErrorMessage)
	}

	return &indResp, nil
}

// SearchSymbol searches for symbols matching the keywords
func (c *Client) SearchSymbol(keywords string) (*SearchResponse, error) {
	params := url.Values{}
	params.Add("function", "SYMBOL_SEARCH")
	params.Add("keywords", keywords)
	params.Add("apikey", c.apiKey)

	resp, err := c.makeRequest(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var searchResp SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling search response: %w", err)
	}

	// Check for API error
	if searchResp.ErrorMessage != "" {
		return nil, fmt.Errorf("alpha vantage API error: %s", searchResp.ErrorMessage)
	}

	return &searchResp, nil
}

// makeRequest makes an HTTP request to the Alpha Vantage API
func (c *Client) makeRequest(params url.Values) (*http.Response, error) {
	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	
	c.logger.Debugf("Making request to: %s", reqURL)
	
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}
	
	return resp, nil
}
