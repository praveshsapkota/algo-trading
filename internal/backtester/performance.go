package backtester

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// PerformanceMetrics represents the performance metrics for a backtest
type PerformanceMetrics struct {
	TotalReturn      float64                `json:"total_return"`
	AnnualizedReturn float64                `json:"annualized_return"`
	WinRate          float64                `json:"win_rate"`
	MaxDrawdown      float64                `json:"max_drawdown"`
	SharpeRatio      float64                `json:"sharpe_ratio"`
	SortinoRatio     float64                `json:"sortino_ratio"`
	TradeCount       int                    `json:"trade_count"`
	ProfitFactor     float64                `json:"profit_factor"`
	AverageTrade     float64                `json:"average_trade"`
	AverageWin       float64                `json:"average_win"`
	AverageLoss      float64                `json:"average_loss"`
	LargestWin       float64                `json:"largest_win"`
	LargestLoss      float64                `json:"largest_loss"`
	AverageHoldTime  string                 `json:"average_hold_time"`
	MonthlyReturns   map[string]float64     `json:"monthly_returns"`
	EquityCurve      []EquityCurvePoint     `json:"equity_curve"`
	DrawdownCurve    []DrawdownCurvePoint   `json:"drawdown_curve"`
	TradeDistribution map[string]int        `json:"trade_distribution"`
}

// EquityCurvePoint represents a point on the equity curve
type EquityCurvePoint struct {
	Date  time.Time `json:"date"`
	Equity float64  `json:"equity"`
}

// DrawdownCurvePoint represents a point on the drawdown curve
type DrawdownCurvePoint struct {
	Date     time.Time `json:"date"`
	Drawdown float64   `json:"drawdown"`
}

// CalculatePerformanceMetrics calculates comprehensive performance metrics for a backtest
func (bs *BacktesterService) CalculatePerformanceMetrics(result *BacktestResult, initialCapital float64) *PerformanceMetrics {
	trades := result.Trades
	if len(trades) == 0 {
		return &PerformanceMetrics{
			TradeCount: 0,
		}
	}

	// Sort trades by entry time
	sort.Slice(trades, func(i, j int) bool {
		return trades[i].EntryTime.Before(trades[j].EntryTime)
	})

	// Calculate basic metrics
	totalPnL := 0.0
	winCount := 0
	lossCount := 0
	totalWin := 0.0
	totalLoss := 0.0
	largestWin := 0.0
	largestLoss := 0.0
	totalHoldTime := 0.0

	for _, trade := range trades {
		totalPnL += trade.PnL

		holdTime := trade.ExitTime.Sub(trade.EntryTime).Hours() / 24 // in days
		totalHoldTime += holdTime

		if trade.PnL > 0 {
			winCount++
			totalWin += trade.PnL
			if trade.PnL > largestWin {
				largestWin = trade.PnL
			}
		} else {
			lossCount++
			totalLoss += math.Abs(trade.PnL)
			if math.Abs(trade.PnL) > largestLoss {
				largestLoss = math.Abs(trade.PnL)
			}
		}
	}

	// Calculate equity curve and drawdown
	equityCurve := make([]EquityCurvePoint, 0)
	drawdownCurve := make([]DrawdownCurvePoint, 0)
	equity := initialCapital
	peak := equity
	maxDrawdown := 0.0

	// Group trades by day for equity curve
	tradesByDay := make(map[string]float64)
	for _, trade := range trades {
		day := trade.ExitTime.Format("2006-01-02")
		tradesByDay[day] += trade.PnL
	}

	// Sort days
	var days []string
	for day := range tradesByDay {
		days = append(days, day)
	}
	sort.Strings(days)

	// Build equity curve and drawdown curve
	for _, day := range days {
		date, _ := time.Parse("2006-01-02", day)
		equity += tradesByDay[day]
		
		equityCurve = append(equityCurve, EquityCurvePoint{
			Date:   date,
			Equity: equity,
		})

		if equity > peak {
			peak = equity
		}
		
		drawdown := (peak - equity) / peak * 100
		drawdownCurve = append(drawdownCurve, DrawdownCurvePoint{
			Date:     date,
			Drawdown: drawdown,
		})

		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	// Calculate monthly returns
	monthlyReturns := make(map[string]float64)
	for _, trade := range trades {
		month := trade.ExitTime.Format("2006-01")
		monthlyReturns[month] += trade.PnL
	}

	// Calculate trade distribution by day of week
	tradeDistribution := make(map[string]int)
	for _, trade := range trades {
		dayOfWeek := trade.EntryTime.Weekday().String()
		tradeDistribution[dayOfWeek]++
	}

	// Calculate annualized return
	firstTradeDate := trades[0].EntryTime
	lastTradeDate := trades[len(trades)-1].ExitTime
	yearFraction := lastTradeDate.Sub(firstTradeDate).Hours() / (24 * 365)
	
	var annualizedReturn float64
	if yearFraction > 0 {
		annualizedReturn = math.Pow((1 + totalPnL/initialCapital), 1/yearFraction) - 1
		annualizedReturn *= 100 // Convert to percentage
	}

	// Calculate Sharpe and Sortino ratios
	dailyReturns := make([]float64, 0)
	for _, pnl := range tradesByDay {
		dailyReturns = append(dailyReturns, pnl/initialCapital)
	}

	sharpeRatio := calculateSharpeRatio(dailyReturns)
	sortinoRatio := calculateSortinoRatio(dailyReturns)

	// Create performance metrics
	metrics := &PerformanceMetrics{
		TotalReturn:      (totalPnL / initialCapital) * 100,
		AnnualizedReturn: annualizedReturn,
		WinRate:          float64(winCount) / float64(len(trades)) * 100,
		MaxDrawdown:      maxDrawdown,
		SharpeRatio:      sharpeRatio,
		SortinoRatio:     sortinoRatio,
		TradeCount:       len(trades),
		ProfitFactor:     calculateProfitFactor(totalWin, totalLoss),
		AverageTrade:     totalPnL / float64(len(trades)),
		AverageWin:       calculateAverage(totalWin, winCount),
		AverageLoss:      calculateAverage(totalLoss, lossCount),
		LargestWin:       largestWin,
		LargestLoss:      largestLoss,
		AverageHoldTime:  formatDuration(totalHoldTime / float64(len(trades))),
		MonthlyReturns:   monthlyReturns,
		EquityCurve:      equityCurve,
		DrawdownCurve:    drawdownCurve,
		TradeDistribution: tradeDistribution,
	}

	return metrics
}

// calculateSharpeRatio calculates the Sharpe ratio from daily returns
func calculateSharpeRatio(returns []float64) float64 {
	if len(returns) < 2 {
		return 0
	}

	// Calculate mean return
	sum := 0.0
	for _, ret := range returns {
		sum += ret
	}
	meanReturn := sum / float64(len(returns))

	// Calculate standard deviation
	sumSquaredDiff := 0.0
	for _, ret := range returns {
		diff := ret - meanReturn
		sumSquaredDiff += diff * diff
	}
	stdDev := math.Sqrt(sumSquaredDiff / float64(len(returns)))

	// Calculate annualized Sharpe ratio (assuming 252 trading days per year)
	if stdDev == 0 {
		return 0
	}
	sharpeRatio := (meanReturn * 252) / (stdDev * math.Sqrt(252))

	return sharpeRatio
}

// calculateSortinoRatio calculates the Sortino ratio from daily returns
func calculateSortinoRatio(returns []float64) float64 {
	if len(returns) < 2 {
		return 0
	}

	// Calculate mean return
	sum := 0.0
	for _, ret := range returns {
		sum += ret
	}
	meanReturn := sum / float64(len(returns))

	// Calculate downside deviation (only negative returns)
	sumSquaredDownside := 0.0
	downsideCount := 0
	for _, ret := range returns {
		if ret < 0 {
			diff := ret - 0 // Target return is 0
			sumSquaredDownside += diff * diff
			downsideCount++
		}
	}

	// Calculate downside deviation
	var downsideDeviation float64
	if downsideCount > 0 {
		downsideDeviation = math.Sqrt(sumSquaredDownside / float64(downsideCount))
	}

	// Calculate annualized Sortino ratio (assuming 252 trading days per year)
	if downsideDeviation == 0 {
		return 0
	}
	sortinoRatio := (meanReturn * 252) / (downsideDeviation * math.Sqrt(252))

	return sortinoRatio
}

// calculateProfitFactor calculates the profit factor
func calculateProfitFactor(totalWin, totalLoss float64) float64 {
	if totalLoss == 0 {
		return 0
	}
	return totalWin / totalLoss
}

// calculateAverage calculates the average value
func calculateAverage(total float64, count int) float64 {
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

// formatDuration formats a duration in days to a human-readable string
func formatDuration(days float64) string {
	if days < 1 {
		hours := days * 24
		return fmt.Sprintf("%.1f hours", hours)
	} else if days < 30 {
		return fmt.Sprintf("%.1f days", days)
	} else {
		months := days / 30
		return fmt.Sprintf("%.1f months", months)
	}
}