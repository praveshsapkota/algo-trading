package options

import (
	"math"
	"time"

	"github.com/user/algo-trading/pkg/models"
)

// BlackScholesModel implements the Black-Scholes option pricing model
type BlackScholesModel struct{}

// PricingResult contains the result of option pricing calculations
type PricingResult struct {
	Price            float64 `json:"price"`
	Delta            float64 `json:"delta"`
	Gamma            float64 `json:"gamma"`
	Theta            float64 `json:"theta"`
	Vega             float64 `json:"vega"`
	ImpliedVolatility float64 `json:"implied_volatility,omitempty"`
}

// NewBlackScholesModel creates a new Black-Scholes pricing model
func NewBlackScholesModel() *BlackScholesModel {
	return &BlackScholesModel{}
}

// CalculatePrice calculates the theoretical price of an option using the Black-Scholes model
func (bs *BlackScholesModel) CalculatePrice(
	optionType models.OptionType,
	underlyingPrice, strikePrice, volatility, riskFreeRate float64,
	timeToExpiration float64,
) *PricingResult {
	// Calculate d1 and d2
	d1 := (math.Log(underlyingPrice/strikePrice) + (riskFreeRate+0.5*volatility*volatility)*timeToExpiration) / (volatility * math.Sqrt(timeToExpiration))
	d2 := d1 - volatility*math.Sqrt(timeToExpiration)

	// Calculate option price
	var price float64
	var delta float64

	if optionType == models.Call {
		price = underlyingPrice*normalCDF(d1) - strikePrice*math.Exp(-riskFreeRate*timeToExpiration)*normalCDF(d2)
		delta = normalCDF(d1)
	} else {
		price = strikePrice*math.Exp(-riskFreeRate*timeToExpiration)*normalCDF(-d2) - underlyingPrice*normalCDF(-d1)
		delta = normalCDF(d1) - 1
	}

	// Calculate Greeks
	gamma := normalPDF(d1) / (underlyingPrice * volatility * math.Sqrt(timeToExpiration))
	vega := underlyingPrice * math.Sqrt(timeToExpiration) * normalPDF(d1) / 100 // Divided by 100 to get the change for a 1% change in volatility
	
	var theta float64
	if optionType == models.Call {
		theta = (-underlyingPrice*normalPDF(d1)*volatility)/(2*math.Sqrt(timeToExpiration)) - 
			riskFreeRate*strikePrice*math.Exp(-riskFreeRate*timeToExpiration)*normalCDF(d2)
	} else {
		theta = (-underlyingPrice*normalPDF(d1)*volatility)/(2*math.Sqrt(timeToExpiration)) + 
			riskFreeRate*strikePrice*math.Exp(-riskFreeRate*timeToExpiration)*normalCDF(-d2)
	}
	// Convert theta to daily value (assuming 365 days in a year)
	theta = theta / 365

	return &PricingResult{
		Price:  price,
		Delta:  delta,
		Gamma:  gamma,
		Theta:  theta,
		Vega:   vega,
	}
}

// CalculateImpliedVolatility calculates the implied volatility of an option
func (bs *BlackScholesModel) CalculateImpliedVolatility(
	optionType models.OptionType,
	underlyingPrice, strikePrice, optionPrice, riskFreeRate float64,
	timeToExpiration float64,
) float64 {
	// Use Newton-Raphson method to find implied volatility
	const maxIterations = 100
	const epsilon = 0.0001

	// Initial guess for volatility
	volatility := 0.3 // 30% as starting point

	for i := 0; i < maxIterations; i++ {
		// Calculate option price with current volatility guess
		result := bs.CalculatePrice(optionType, underlyingPrice, strikePrice, volatility, riskFreeRate, timeToExpiration)
		price := result.Price
		vega := result.Vega

		// Calculate difference between calculated price and market price
		diff := price - optionPrice

		// Check if we're close enough
		if math.Abs(diff) < epsilon {
			return volatility
		}

		// Update volatility using Newton-Raphson step
		if vega == 0 {
			break // Avoid division by zero
		}
		volatility = volatility - diff/vega
		
		// Ensure volatility stays within reasonable bounds
		if volatility <= 0.001 {
			volatility = 0.001
		} else if volatility > 5 {
			volatility = 5 // Cap at 500%
		}
	}

	return volatility
}

// CalculateTimeToExpiration calculates the time to expiration in years
func CalculateTimeToExpiration(expirationDate time.Time) float64 {
	now := time.Now()
	if now.After(expirationDate) {
		return 0
	}
	
	// Calculate time to expiration in years (assuming 365 days in a year)
	return expirationDate.Sub(now).Hours() / (24 * 365)
}

// normalCDF calculates the cumulative distribution function for a standard normal distribution
func normalCDF(x float64) float64 {
	return 0.5 * (1 + math.Erf(x/math.Sqrt(2)))
}

// normalPDF calculates the probability density function for a standard normal distribution
func normalPDF(x float64) float64 {
	return math.Exp(-0.5*x*x) / math.Sqrt(2*math.Pi)
}