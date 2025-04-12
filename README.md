# High-Frequency Algorithmic Trading System

A high-performance, scalable algorithmic trading system for stocks using Go, with Alpha Vantage/Yahoo Finance as data sources and PostgreSQL for data storage.

## System Architecture

The system is built using a microservices architecture with the following components:

1. **Market Data Service**: Connects to Alpha Vantage and Yahoo Finance APIs to fetch real-time and historical market data.
2. **Data Processing Engine**: Processes and normalizes market data for analysis.
3. **Signal Generator**: Applies trading strategies to generate buy/sell signals.
4. **Trade Validator**: Validates signals and checks for duplicate trades.
5. **Risk Manager**: Evaluates risk/reward ratios and enforces risk limits.
6. **Order Executor**: Handles order placement and execution.
7. **Trade Repository**: Stores trade information in PostgreSQL.
8. **Dashboard Service**: Provides a web interface for monitoring and analysis.

## Technology Stack

- **Backend**: Go (Gin, GORM, gorilla/websocket)
- **Data Processing**: Redis, InfluxDB
- **Storage**: PostgreSQL, TimescaleDB
- **Frontend**: React, D3.js, TradingView Charts
- **DevOps**: Docker, Kubernetes, Prometheus, Grafana

## Getting Started

### Prerequisites

- Go 1.20+
- Docker and Docker Compose
- PostgreSQL 14+
- Redis

### Installation

1. Clone the repository:
   ```
   git clone https://github.com/user/algo-trading.git
   cd algo-trading
   ```

2. Set up environment variables:
   ```
   cp configs/example.env configs/.env
   # Edit .env file with your API keys and configuration
   ```

3. Build and run with Docker Compose:
   ```
   docker-compose up -d
   ```

4. Access the dashboard:
   ```
   http://localhost:3000
   ```

## Project Structure

```
algo-trading/
├── cmd/                    # Application entry points
│   ├── market-data/        # Market data service
│   ├── signal-generator/   # Signal generation service
│   ├── trade-executor/     # Trade execution service
│   ├── risk-manager/       # Risk management service
│   └── dashboard/          # Dashboard service
├── pkg/                    # Shared packages
│   ├── models/             # Data models
│   ├── database/           # Database access
│   ├── api/                # API clients
│   ├── strategy/           # Trading strategies
│   ├── indicators/         # Technical indicators
│   ├── risk/               # Risk management
│   └── utils/              # Utility functions
├── internal/               # Internal packages
│   ├── market-data/        # Market data implementation
│   ├── signal-generator/   # Signal generation implementation
│   ├── trade-executor/     # Trade execution implementation
│   ├── risk-manager/       # Risk management implementation
│   └── dashboard/          # Dashboard implementation
├── api/                    # API definitions
│   ├── proto/              # Protocol buffers
│   └── rest/               # REST API specs
├── configs/                # Configuration files
├── deployments/            # Deployment configurations
│   ├── docker/             # Docker configurations
│   └── kubernetes/         # Kubernetes manifests
├── web/                    # Web dashboard
│   ├── src/                # Frontend source code
│   └── public/             # Static assets
├── scripts/                # Utility scripts
└── docs/                   # Documentation
```

## Development

### Running Services Locally

To run individual services locally:

```bash
# Market Data Service
cd cmd/market-data
go run main.go

# Signal Generator
cd cmd/signal-generator
go run main.go

# Dashboard
cd cmd/dashboard
go run main.go
```

### Testing

```bash
go test ./...
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
