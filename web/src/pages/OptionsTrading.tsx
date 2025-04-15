import React, { useState, useEffect } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  Container,
  FormControl,
  Grid,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Slider,
  Tab,
  Tabs,
  TextField,
  Typography,
} from '@mui/material';
import { DatePicker } from '@mui/x-date-pickers/DatePicker';
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';

interface OptionChain {
  underlyingSymbol: string;
  underlyingPrice: number;
  updatedAt: string;
  calls: Option[];
  puts: Option[];
}

interface Option {
  id: number;
  underlyingSymbol: string;
  type: 'CALL' | 'PUT';
  strikePrice: number;
  expirationDate: string;
  bid: number;
  ask: number;
  lastPrice: number;
  volume: number;
  openInterest: number;
  impliedVolatility: number;
  delta: number;
  gamma: number;
  theta: number;
  vega: number;
}

interface OptionPosition {
  id: number;
  underlyingSymbol: string;
  type: 'CALL' | 'PUT';
  strikePrice: number;
  expirationDate: string;
  quantity: number;
  averagePrice: number;
  currentPrice: number;
  pnl: number;
}

const OptionsTrading: React.FC = () => {
  const [symbol, setSymbol] = useState<string>('AAPL');
  const [expirationDates, setExpirationDates] = useState<string[]>([]);
  const [selectedExpiration, setSelectedExpiration] = useState<string>('');
  const [tabValue, setTabValue] = useState<number>(0);
  const [optionChain, setOptionChain] = useState<OptionChain | null>(null);
  const [positions, setPositions] = useState<OptionPosition[]>([]);
  const [underlyingPrice, setUnderlyingPrice] = useState<number>(0);
  const [calculatorParams, setCalculatorParams] = useState({
    type: 'CALL',
    strikePrice: 100,
    expirationDate: new Date(),
    volatility: 0.3,
    riskFreeRate: 0.03,
  });

  // Fetch option chain data
  useEffect(() => {
    const fetchOptionChain = async () => {
      try {
        // In a real implementation, this would fetch from the API
        // For now, we'll use mock data
        const mockData: OptionChain = {
          underlyingSymbol: symbol,
          underlyingPrice: 150.25,
          updatedAt: new Date().toISOString(),
          calls: generateMockOptions(symbol, 'CALL', 150.25),
          puts: generateMockOptions(symbol, 'PUT', 150.25),
        };
        
        setOptionChain(mockData);
        setUnderlyingPrice(mockData.underlyingPrice);
        
        // Extract expiration dates
        const dates = [...new Set([
          ...mockData.calls.map(option => option.expirationDate),
          ...mockData.puts.map(option => option.expirationDate),
        ])].sort();
        
        setExpirationDates(dates);
        if (dates.length > 0 && !selectedExpiration) {
          setSelectedExpiration(dates[0]);
        }
      } catch (error) {
        console.error('Error fetching option chain:', error);
      }
    };

    fetchOptionChain();
  }, [symbol]);

  // Fetch positions
  useEffect(() => {
    const fetchPositions = async () => {
      try {
        // In a real implementation, this would fetch from the API
        // For now, we'll use mock data
        const mockPositions: OptionPosition[] = [
          {
            id: 1,
            underlyingSymbol: 'AAPL',
            type: 'CALL',
            strikePrice: 155,
            expirationDate: '2025-06-20',
            quantity: 2,
            averagePrice: 5.75,
            currentPrice: 6.25,
            pnl: 100,
          },
          {
            id: 2,
            underlyingSymbol: 'AAPL',
            type: 'PUT',
            strikePrice: 145,
            expirationDate: '2025-06-20',
            quantity: 1,
            averagePrice: 4.50,
            currentPrice: 3.75,
            pnl: -75,
          },
        ];
        
        setPositions(mockPositions);
      } catch (error) {
        console.error('Error fetching positions:', error);
      }
    };

    fetchPositions();
  }, []);

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  const handleSymbolChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setSymbol(event.target.value.toUpperCase());
  };

  const handleExpirationChange = (event: any) => {
    setSelectedExpiration(event.target.value);
  };

  const handleCalculatorParamChange = (param: string, value: any) => {
    setCalculatorParams(prev => ({
      ...prev,
      [param]: value
    }));
  };

  // Generate mock options data
  const generateMockOptions = (symbol: string, type: 'CALL' | 'PUT', price: number): Option[] => {
    const options: Option[] = [];
    const today = new Date();
    
    // Generate 3 expiration dates
    const expirations = [
      new Date(today.getFullYear(), today.getMonth() + 1, 15).toISOString().split('T')[0],
      new Date(today.getFullYear(), today.getMonth() + 2, 15).toISOString().split('T')[0],
      new Date(today.getFullYear(), today.getMonth() + 3, 15).toISOString().split('T')[0],
    ];
    
    // Generate strikes around the current price
    const strikes = [
      price * 0.9,
      price * 0.95,
      price,
      price * 1.05,
      price * 1.1,
    ].map(p => Math.round(p));
    
    let id = 1;
    
    for (const expiration of expirations) {
      for (const strike of strikes) {
        // Calculate mock option price
        const timeToExpiration = (new Date(expiration).getTime() - today.getTime()) / (365 * 24 * 60 * 60 * 1000);
        const volatility = 0.3;
        let optionPrice;
        
        if (type === 'CALL') {
          optionPrice = Math.max(0, price - strike) + (strike * volatility * Math.sqrt(timeToExpiration));
        } else {
          optionPrice = Math.max(0, strike - price) + (strike * volatility * Math.sqrt(timeToExpiration));
        }
        
        options.push({
          id: id++,
          underlyingSymbol: symbol,
          type,
          strikePrice: strike,
          expirationDate: expiration,
          bid: parseFloat((optionPrice * 0.95).toFixed(2)),
          ask: parseFloat((optionPrice * 1.05).toFixed(2)),
          lastPrice: parseFloat(optionPrice.toFixed(2)),
          volume: Math.floor(Math.random() * 1000),
          openInterest: Math.floor(Math.random() * 5000),
          impliedVolatility: volatility,
          delta: type === 'CALL' ? 0.5 + (price - strike) / (price * 2) : 0.5 - (price - strike) / (price * 2),
          gamma: 0.05,
          theta: -0.03,
          vega: 0.1,
        });
      }
    }
    
    return options;
  };

  // Calculate option price using Black-Scholes (simplified)
  const calculateOptionPrice = () => {
    const { type, strikePrice, expirationDate, volatility, riskFreeRate } = calculatorParams;
    
    // Calculate time to expiration in years
    const timeToExpiration = (expirationDate.getTime() - new Date().getTime()) / (365 * 24 * 60 * 60 * 1000);
    
    // Simplified Black-Scholes calculation
    let optionPrice;
    if (type === 'CALL') {
      optionPrice = Math.max(0, underlyingPrice - strikePrice) + 
        (strikePrice * volatility * Math.sqrt(timeToExpiration));
    } else {
      optionPrice = Math.max(0, strikePrice - underlyingPrice) + 
        (strikePrice * volatility * Math.sqrt(timeToExpiration));
    }
    
    return optionPrice.toFixed(2);
  };

  return (
    <LocalizationProvider dateAdapter={AdapterDateFns}>
      <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
        <Typography variant="h4" gutterBottom>
          Options Trading
        </Typography>
        
        <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 3 }}>
          <Tabs value={tabValue} onChange={handleTabChange}>
            <Tab label="Option Chain" />
            <Tab label="My Positions" />
            <Tab label="Option Calculator" />
          </Tabs>
        </Box>
        
        {/* Option Chain Tab */}
        {tabValue === 0 && (
          <Box>
            <Grid container spacing={3} sx={{ mb: 3 }}>
              <Grid item xs={12} md={4}>
                <TextField
                  label="Symbol"
                  value={symbol}
                  onChange={handleSymbolChange}
                  fullWidth
                />
              </Grid>
              <Grid item xs={12} md={4}>
                <FormControl fullWidth>
                  <InputLabel>Expiration Date</InputLabel>
                  <Select
                    value={selectedExpiration}
                    label="Expiration Date"
                    onChange={handleExpirationChange}
                  >
                    {expirationDates.map((date) => (
                      <MenuItem key={date} value={date}>
                        {new Date(date).toLocaleDateString()}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </Grid>
              <Grid item xs={12} md={4}>
                <Typography variant="h6" align="right">
                  {symbol}: ${underlyingPrice.toFixed(2)}
                </Typography>
              </Grid>
            </Grid>
            
            <Grid container spacing={3}>
              {/* Calls */}
              <Grid item xs={12} md={6}>
                <Paper sx={{ p: 2 }}>
                  <Typography variant="h6" gutterBottom>
                    Calls
                  </Typography>
                  <Box sx={{ overflowX: 'auto' }}>
                    <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                      <thead>
                        <tr>
                          <th>Strike</th>
                          <th>Bid</th>
                          <th>Ask</th>
                          <th>Last</th>
                          <th>IV</th>
                          <th>Volume</th>
                        </tr>
                      </thead>
                      <tbody>
                        {optionChain?.calls
                          .filter(option => option.expirationDate === selectedExpiration)
                          .map(option => (
                            <tr key={option.id}>
                              <td>{option.strikePrice.toFixed(2)}</td>
                              <td>{option.bid.toFixed(2)}</td>
                              <td>{option.ask.toFixed(2)}</td>
                              <td>{option.lastPrice.toFixed(2)}</td>
                              <td>{(option.impliedVolatility * 100).toFixed(1)}%</td>
                              <td>{option.volume}</td>
                            </tr>
                          ))}
                      </tbody>
                    </table>
                  </Box>
                </Paper>
              </Grid>
              
              {/* Puts */}
              <Grid item xs={12} md={6}>
                <Paper sx={{ p: 2 }}>
                  <Typography variant="h6" gutterBottom>
                    Puts
                  </Typography>
                  <Box sx={{ overflowX: 'auto' }}>
                    <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                      <thead>
                        <tr>
                          <th>Strike</th>
                          <th>Bid</th>
                          <th>Ask</th>
                          <th>Last</th>
                          <th>IV</th>
                          <th>Volume</th>
                        </tr>
                      </thead>
                      <tbody>
                        {optionChain?.puts
                          .filter(option => option.expirationDate === selectedExpiration)
                          .map(option => (
                            <tr key={option.id}>
                              <td>{option.strikePrice.toFixed(2)}</td>
                              <td>{option.bid.toFixed(2)}</td>
                              <td>{option.ask.toFixed(2)}</td>
                              <td>{option.lastPrice.toFixed(2)}</td>
                              <td>{(option.impliedVolatility * 100).toFixed(1)}%</td>
                              <td>{option.volume}</td>
                            </tr>
                          ))}
                      </tbody>
                    </table>
                  </Box>
                </Paper>
              </Grid>
            </Grid>
          </Box>
        )}
        
        {/* Positions Tab */}
        {tabValue === 1 && (
          <Box>
            <Paper sx={{ p: 2 }}>
              <Typography variant="h6" gutterBottom>
                Open Positions
              </Typography>
              <Box sx={{ overflowX: 'auto' }}>
                <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                  <thead>
                    <tr>
                      <th>Symbol</th>
                      <th>Type</th>
                      <th>Strike</th>
                      <th>Expiration</th>
                      <th>Quantity</th>
                      <th>Avg Price</th>
                      <th>Current</th>
                      <th>P&L</th>
                    </tr>
                  </thead>
                  <tbody>
                    {positions.map(position => (
                      <tr key={position.id}>
                        <td>{position.underlyingSymbol}</td>
                        <td>{position.type}</td>
                        <td>${position.strikePrice.toFixed(2)}</td>
                        <td>{new Date(position.expirationDate).toLocaleDateString()}</td>
                        <td>{position.quantity}</td>
                        <td>${position.averagePrice.toFixed(2)}</td>
                        <td>${position.currentPrice.toFixed(2)}</td>
                        <td style={{ color: position.pnl >= 0 ? 'green' : 'red' }}>
                          ${position.pnl.toFixed(2)}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </Box>
            </Paper>
          </Box>
        )}
        
        {/* Option Calculator Tab */}
        {tabValue === 2 && (
          <Box>
            <Paper sx={{ p: 2 }}>
              <Typography variant="h6" gutterBottom>
                Option Price Calculator
              </Typography>
              <Grid container spacing={3}>
                <Grid item xs={12} md={6}>
                  <FormControl fullWidth sx={{ mb: 2 }}>
                    <InputLabel>Option Type</InputLabel>
                    <Select
                      value={calculatorParams.type}
                      label="Option Type"
                      onChange={(e) => handleCalculatorParamChange('type', e.target.value)}
                    >
                      <MenuItem value="CALL">Call</MenuItem>
                      <MenuItem value="PUT">Put</MenuItem>
                    </Select>
                  </FormControl>
                  
                  <TextField
                    label="Underlying Price"
                    type="number"
                    value={underlyingPrice}
                    onChange={(e) => setUnderlyingPrice(parseFloat(e.target.value))}
                    fullWidth
                    sx={{ mb: 2 }}
                  />
                  
                  <TextField
                    label="Strike Price"
                    type="number"
                    value={calculatorParams.strikePrice}
                    onChange={(e) => handleCalculatorParamChange('strikePrice', parseFloat(e.target.value))}
                    fullWidth
                    sx={{ mb: 2 }}
                  />
                  
                  <DatePicker
                    label="Expiration Date"
                    value={calculatorParams.expirationDate}
                    onChange={(date) => date && handleCalculatorParamChange('expirationDate', date)}
                    sx={{ mb: 2, width: '100%' }}
                  />
                  
                  <Typography gutterBottom>
                    Implied Volatility: {(calculatorParams.volatility * 100).toFixed(1)}%
                  </Typography>
                  <Slider
                    value={calculatorParams.volatility * 100}
                    onChange={(e, value) => handleCalculatorParamChange('volatility', (value as number) / 100)}
                    min={5}
                    max={100}
                    step={1}
                    valueLabelDisplay="auto"
                    valueLabelFormat={(value) => `${value}%`}
                    sx={{ mb: 2 }}
                  />
                  
                  <Typography gutterBottom>
                    Risk-Free Rate: {(calculatorParams.riskFreeRate * 100).toFixed(1)}%
                  </Typography>
                  <Slider
                    value={calculatorParams.riskFreeRate * 100}
                    onChange={(e, value) => handleCalculatorParamChange('riskFreeRate', (value as number) / 100)}
                    min={0}
                    max={10}
                    step={0.1}
                    valueLabelDisplay="auto"
                    valueLabelFormat={(value) => `${value}%`}
                    sx={{ mb: 2 }}
                  />
                </Grid>
                
                <Grid item xs={12} md={6}>
                  <Card sx={{ height: '100%', display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                    <CardContent>
                      <Typography variant="h5" align="center" gutterBottom>
                        Theoretical Price
                      </Typography>
                      <Typography variant="h3" align="center" color="primary">
                        ${calculateOptionPrice()}
                      </Typography>
                      
                      <Box sx={{ mt: 4 }}>
                        <Typography variant="subtitle1" gutterBottom>
                          Greeks:
                        </Typography>
                        <Grid container spacing={2}>
                          <Grid item xs={6}>
                            <Typography>
                              Delta: {calculatorParams.type === 'CALL' ? '+' : '-'}0.65
                            </Typography>
                          </Grid>
                          <Grid item xs={6}>
                            <Typography>
                              Gamma: 0.05
                            </Typography>
                          </Grid>
                          <Grid item xs={6}>
                            <Typography>
                              Theta: -0.03
                            </Typography>
                          </Grid>
                          <Grid item xs={6}>
                            <Typography>
                              Vega: 0.10
                            </Typography>
                          </Grid>
                        </Grid>
                      </Box>
                    </CardContent>
                  </Card>
                </Grid>
              </Grid>
            </Paper>
          </Box>
        )}
      </Container>
    </LocalizationProvider>
  );
};

export default OptionsTrading;