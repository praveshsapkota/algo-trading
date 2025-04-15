import React, { useState, useEffect } from 'react';
import { 
  Box, 
  Button, 
  Card, 
  CardContent, 
  CircularProgress, 
  Container, 
  FormControl, 
  Grid, 
  InputLabel, 
  MenuItem, 
  Paper, 
  Select, 
  TextField, 
  Typography 
} from '@mui/material';
import { DatePicker } from '@mui/x-date-pickers/DatePicker';
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { 
  LineChart, 
  Line, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  Legend, 
  ResponsiveContainer,
  AreaChart,
  Area
} from 'recharts';

interface BacktestParams {
  strategyName: string;
  symbols: string[];
  startDate: Date;
  endDate: Date;
  parameters: Record<string, any>;
  timeframe: string;
}

interface BacktestResult {
  id: number;
  strategyName: string;
  symbols: string[];
  startDate: string;
  endDate: string;
  parameters: Record<string, any>;
  totalReturn: number;
  winRate: number;
  maxDrawdown: number;
  sharpeRatio: number;
  tradeCount: number;
  trades: any[];
  createdAt: string;
}

const Backtester: React.FC = () => {
  const [strategies, setStrategies] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [results, setResults] = useState<BacktestResult | null>(null);
  const [equityCurve, setEquityCurve] = useState<any[]>([]);
  const [drawdownCurve, setDrawdownCurve] = useState<any[]>([]);
  const [params, setParams] = useState<BacktestParams>({
    strategyName: '',
    symbols: ['AAPL'],
    startDate: new Date(new Date().setFullYear(new Date().getFullYear() - 1)),
    endDate: new Date(),
    parameters: {},
    timeframe: '1d'
  });

  // Fetch available strategies
  useEffect(() => {
    const fetchStrategies = async () => {
      try {
        const response = await fetch('/api/backtest/strategies');
        const data = await response.json();
        setStrategies(data);
        if (data.length > 0) {
          setParams(prev => ({ ...prev, strategyName: data[0] }));
        }
      } catch (error) {
        console.error('Error fetching strategies:', error);
      }
    };

    fetchStrategies();
  }, []);

  const handleInputChange = (field: keyof BacktestParams, value: any) => {
    setParams(prev => ({ ...prev, [field]: value }));
  };

  const handleSymbolChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const symbols = event.target.value.split(',').map(s => s.trim()).filter(s => s);
    setParams(prev => ({ ...prev, symbols }));
  };

  const handleParameterChange = (key: string, value: any) => {
    setParams(prev => ({
      ...prev,
      parameters: {
        ...prev.parameters,
        [key]: value
      }
    }));
  };

  const runBacktest = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/backtest/run', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(params)
      });
      
      const result = await response.json();
      setResults(result);
      
      // Generate equity curve data
      if (result.trades && result.trades.length > 0) {
        generateEquityCurve(result.trades);
      }
    } catch (error) {
      console.error('Error running backtest:', error);
    } finally {
      setLoading(false);
    }
  };

  const generateEquityCurve = (trades: any[]) => {
    // Sort trades by date
    const sortedTrades = [...trades].sort((a, b) => 
      new Date(a.exitTime).getTime() - new Date(b.exitTime).getTime()
    );

    // Group trades by day
    const tradesByDay: Record<string, number> = {};
    sortedTrades.forEach(trade => {
      const date = new Date(trade.exitTime).toISOString().split('T')[0];
      tradesByDay[date] = (tradesByDay[date] || 0) + trade.pnl;
    });

    // Generate equity curve
    const initialCapital = 100000; // Assuming $100k starting capital
    let equity = initialCapital;
    let peak = equity;
    
    const equityData: any[] = [];
    const drawdownData: any[] = [];

    Object.entries(tradesByDay).forEach(([date, pnl]) => {
      equity += pnl;
      equityData.push({
        date,
        equity
      });

      // Calculate drawdown
      peak = Math.max(peak, equity);
      const drawdown = ((peak - equity) / peak) * 100;
      drawdownData.push({
        date,
        drawdown
      });
    });

    setEquityCurve(equityData);
    setDrawdownCurve(drawdownData);
  };

  return (
    <LocalizationProvider dateAdapter={AdapterDateFns}>
      <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
        <Typography variant="h4" gutterBottom>
          Strategy Backtester
        </Typography>
        
        <Grid container spacing={3}>
          {/* Backtest Parameters */}
          <Grid item xs={12} md={4}>
            <Paper sx={{ p: 2, display: 'flex', flexDirection: 'column' }}>
              <Typography variant="h6" gutterBottom>
                Parameters
              </Typography>
              
              <FormControl fullWidth sx={{ mb: 2 }}>
                <InputLabel>Strategy</InputLabel>
                <Select
                  value={params.strategyName}
                  label="Strategy"
                  onChange={(e) => handleInputChange('strategyName', e.target.value)}
                >
                  {strategies.map(strategy => (
                    <MenuItem key={strategy} value={strategy}>{strategy}</MenuItem>
                  ))}
                </Select>
              </FormControl>
              
              <TextField
                label="Symbols (comma separated)"
                value={params.symbols.join(', ')}
                onChange={handleSymbolChange}
                fullWidth
                margin="normal"
              />
              
              <FormControl fullWidth sx={{ mb: 2, mt: 2 }}>
                <InputLabel>Timeframe</InputLabel>
                <Select
                  value={params.timeframe}
                  label="Timeframe"
                  onChange={(e) => handleInputChange('timeframe', e.target.value)}
                >
                  <MenuItem value="1m">1 Minute</MenuItem>
                  <MenuItem value="5m">5 Minutes</MenuItem>
                  <MenuItem value="15m">15 Minutes</MenuItem>
                  <MenuItem value="1h">1 Hour</MenuItem>
                  <MenuItem value="1d">1 Day</MenuItem>
                </Select>
              </FormControl>
              
              <DatePicker
                label="Start Date"
                value={params.startDate}
                onChange={(date) => date && handleInputChange('startDate', date)}
                sx={{ mb: 2 }}
              />
              
              <DatePicker
                label="End Date"
                value={params.endDate}
                onChange={(date) => date && handleInputChange('endDate', date)}
                sx={{ mb: 2 }}
              />
              
              <Button 
                variant="contained" 
                color="primary" 
                onClick={runBacktest}
                disabled={loading || !params.strategyName || params.symbols.length === 0}
                sx={{ mt: 2 }}
              >
                {loading ? <CircularProgress size={24} /> : 'Run Backtest'}
              </Button>
            </Paper>
          </Grid>
          
          {/* Results */}
          <Grid item xs={12} md={8}>
            {results ? (
              <Box>
                <Paper sx={{ p: 2, mb: 2 }}>
                  <Typography variant="h6" gutterBottom>
                    Backtest Results
                  </Typography>
                  
                  <Grid container spacing={2}>
                    <Grid item xs={6} md={3}>
                      <Card>
                        <CardContent>
                          <Typography variant="subtitle2" color="text.secondary">
                            Total Return
                          </Typography>
                          <Typography variant="h6" color={results.totalReturn >= 0 ? 'success.main' : 'error.main'}>
                            {results.totalReturn.toFixed(2)}%
                          </Typography>
                        </CardContent>
                      </Card>
                    </Grid>
                    
                    <Grid item xs={6} md={3}>
                      <Card>
                        <CardContent>
                          <Typography variant="subtitle2" color="text.secondary">
                            Win Rate
                          </Typography>
                          <Typography variant="h6">
                            {results.winRate.toFixed(2)}%
                          </Typography>
                        </CardContent>
                      </Card>
                    </Grid>
                    
                    <Grid item xs={6} md={3}>
                      <Card>
                        <CardContent>
                          <Typography variant="subtitle2" color="text.secondary">
                            Max Drawdown
                          </Typography>
                          <Typography variant="h6" color="error.main">
                            {results.maxDrawdown.toFixed(2)}%
                          </Typography>
                        </CardContent>
                      </Card>
                    </Grid>
                    
                    <Grid item xs={6} md={3}>
                      <Card>
                        <CardContent>
                          <Typography variant="subtitle2" color="text.secondary">
                            Sharpe Ratio
                          </Typography>
                          <Typography variant="h6">
                            {results.sharpeRatio.toFixed(2)}
                          </Typography>
                        </CardContent>
                      </Card>
                    </Grid>
                  </Grid>
                </Paper>
                
                {/* Equity Curve */}
                {equityCurve.length > 0 && (
                  <Paper sx={{ p: 2, mb: 2, height: 300 }}>
                    <Typography variant="h6" gutterBottom>
                      Equity Curve
                    </Typography>
                    <ResponsiveContainer width="100%" height="100%">
                      <LineChart
                        data={equityCurve}
                        margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
                      >
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis dataKey="date" />
                        <YAxis />
                        <Tooltip />
                        <Legend />
                        <Line 
                          type="monotone" 
                          dataKey="equity" 
                          stroke="#8884d8" 
                          activeDot={{ r: 8 }} 
                        />
                      </LineChart>
                    </ResponsiveContainer>
                  </Paper>
                )}
                
                {/* Drawdown Chart */}
                {drawdownCurve.length > 0 && (
                  <Paper sx={{ p: 2, mb: 2, height: 300 }}>
                    <Typography variant="h6" gutterBottom>
                      Drawdown
                    </Typography>
                    <ResponsiveContainer width="100%" height="100%">
                      <AreaChart
                        data={drawdownCurve}
                        margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
                      >
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis dataKey="date" />
                        <YAxis />
                        <Tooltip />
                        <Legend />
                        <Area 
                          type="monotone" 
                          dataKey="drawdown" 
                          stroke="#ff8042" 
                          fill="#ff8042" 
                        />
                      </AreaChart>
                    </ResponsiveContainer>
                  </Paper>
                )}
              </Box>
            ) : (
              <Paper sx={{ p: 2, display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
                <Typography variant="subtitle1" color="text.secondary">
                  Run a backtest to see results
                </Typography>
              </Paper>
            )}
          </Grid>
        </Grid>
      </Container>
    </LocalizationProvider>
  );
};

export default Backtester;