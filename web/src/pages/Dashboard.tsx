import React, { useEffect, useState } from 'react';
import { Grid, Paper, Typography } from '@mui/material';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

import StatsCard from '../components/StatsCard';
import TradeList from '../components/TradeList';
import SignalList from '../components/SignalList';

interface DashboardStats {
  totalTrades: number;
  winningTrades: number;
  losingTrades: number;
  totalPnl: number;
  winRate: number;
  averagePnl: number;
  openPositions: number;
  dailyPnl: number;
  falseSignals: number;
  signalAccuracy: number;
}

const Dashboard: React.FC = () => {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [pnlHistory, setPnlHistory] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        // Fetch dashboard statistics
        const statsResponse = await fetch('/api/stats');
        const statsData = await statsResponse.json();
        setStats(statsData);

        // Fetch PnL history for the chart
        const pnlResponse = await fetch('/api/history/pnl');
        const pnlData = await pnlResponse.json();
        setPnlHistory(pnlData);
      } catch (error) {
        console.error('Error fetching dashboard data:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
    const interval = setInterval(fetchData, 60000); // Refresh every minute

    return () => clearInterval(interval);
  }, []);

  if (loading || !stats) {
    return <Typography>Loading...</Typography>;
  }

  return (
    <Grid container spacing={3}>
      {/* Statistics Cards */}
      <Grid item xs={12} md={3}>
        <StatsCard
          title="Total PnL"
          value={`$${stats.totalPnl.toFixed(2)}`}
          color={stats.totalPnl >= 0 ? 'success' : 'error'}
        />
      </Grid>
      <Grid item xs={12} md={3}>
        <StatsCard
          title="Win Rate"
          value={`${stats.winRate.toFixed(1)}%`}
          color="primary"
        />
      </Grid>
      <Grid item xs={12} md={3}>
        <StatsCard
          title="Open Positions"
          value={stats.openPositions.toString()}
          color="info"
        />
      </Grid>
      <Grid item xs={12} md={3}>
        <StatsCard
          title="Signal Accuracy"
          value={`${stats.signalAccuracy.toFixed(1)}%`}
          color="secondary"
        />
      </Grid>

      {/* PnL Chart */}
      <Grid item xs={12}>
        <Paper sx={{ p: 2, height: 400 }}>
          <Typography variant="h6" gutterBottom>
            PnL History
          </Typography>
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart
              data={pnlHistory}
              margin={{ top: 10, right: 30, left: 0, bottom: 0 }}
            >
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="date" />
              <YAxis />
              <Tooltip />
              <Area
                type="monotone"
                dataKey="pnl"
                stroke="#8884d8"
                fill="#8884d8"
                fillOpacity={0.3}
              />
            </AreaChart>
          </ResponsiveContainer>
        </Paper>
      </Grid>

      {/* Recent Trades */}
      <Grid item xs={12} md={6}>
        <Paper sx={{ p: 2, height: 400 }}>
          <Typography variant="h6" gutterBottom>
            Recent Trades
          </Typography>
          <TradeList limit={5} />
        </Paper>
      </Grid>

      {/* Active Signals */}
      <Grid item xs={12} md={6}>
        <Paper sx={{ p: 2, height: 400 }}>
          <Typography variant="h6" gutterBottom>
            Active Signals
          </Typography>
          <SignalList limit={5} />
        </Paper>
      </Grid>

      {/* Additional Statistics */}
      <Grid item xs={12} md={4}>
        <StatsCard
          title="Total Trades"
          value={stats.totalTrades.toString()}
          color="default"
        />
      </Grid>
      <Grid item xs={12} md={4}>
        <StatsCard
          title="Average PnL"
          value={`$${stats.averagePnl.toFixed(2)}`}
          color={stats.averagePnl >= 0 ? 'success' : 'error'}
        />
      </Grid>
      <Grid item xs={12} md={4}>
        <StatsCard
          title="False Signals"
          value={stats.falseSignals.toString()}
          color="warning"
        />
      </Grid>
    </Grid>
  );
};

export default Dashboard;