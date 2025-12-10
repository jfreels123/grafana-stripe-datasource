# Stripe Data Source for Grafana

A Grafana data source plugin for visualizing Stripe billing metrics including MRR, ARR, churn rate, and more.

## Features

### Revenue Metrics
- **MRR** - Monthly Recurring Revenue
- **ARR** - Annual Recurring Revenue (MRR × 12)
- **New MRR** - MRR from subscriptions created in the last 30 days
- **Churned MRR** - MRR lost from canceled subscriptions (last 30 days)
- **Net New MRR** - New MRR minus Churned MRR
- **Total Revenue** - Sum of all paid invoices
- **ARPU** - Average Revenue Per User

### Subscriber Metrics
- **Active Subscribers** - Count of active subscriptions
- **Churn Rate** - Subscriber churn percentage (last 30 days)
- **Trialing** - Subscriptions currently in trial
- **Past Due** - Subscriptions with overdue payments
- **Total Customers** - Customer count

### Data Tables
- **Subscriptions** - All active subscriptions with details
- **Invoices** - Invoice history with status and amounts
- **Charges** - Payment charges with success/failure status
- **Revenue by Product** - MRR breakdown by product

### Other
- **Available Balance** - USD balance available for payout

## Requirements

- Grafana 10.4.0+
- Stripe API key with read permissions

## Installation

1. Download the latest release from [GitHub Releases](https://github.com/jfreels123/grafana-stripe-datasource/releases)
2. Extract to your Grafana plugins directory (e.g., `/var/lib/grafana/plugins/`)
3. For unsigned plugins, add to `grafana.ini`:
   ```ini
   [plugins]
   allow_loading_unsigned_plugins = jfreels123-stripe-datasource
   ```
4. Restart Grafana

## Configuration

1. Go to **Connections → Data sources → Add data source**
2. Search for "Stripe"
3. Enter your Stripe API key
4. Click **Save & test**

### API Key Setup

Create a [restricted API key](https://dashboard.stripe.com/apikeys) in Stripe Dashboard:

1. Go to **Developers → API keys → Create restricted key**
2. Select **"Building your own integration"**
3. Set the following permissions to **Read**:

| Resource | Permission | Used For |
|----------|------------|----------|
| Customers | Read | Customer count |
| Subscriptions | Read | MRR, ARR, subscriber metrics |
| Balance | Read | Available balance |
| Invoices | Read | Revenue, invoice table |
| Charges | Read | Charges table |
| Products | Read | Revenue by product |
| Prices | Read | Product pricing details |

4. Click **Create key**
5. Copy the key (starts with `rk_live_...` or `rk_test_...`)

You can also use a standard secret key (`sk_live_...` or `sk_test_...`) but restricted keys are recommended for security.

## Usage

### Single Value Metrics (Stat/Gauge panels)

Select any metric like MRR, ARR, Churn Rate, etc. Best displayed with **Stat** or **Gauge** visualization.

### Data Tables

Select Subscriptions, Invoices, Charges, or Revenue by Product. Use **Table** visualization.

### Dashboard Example

Create a dashboard with:
- Stat panels for MRR, ARR, Active Subscribers, Churn Rate
- Table panel for Subscriptions list
- Stat panel for Available Balance

### Limitations

- **Current state only**: Stripe API returns current values, not historical data. For time-series charts showing MRR over time, you'll need to store snapshots in a separate database.
- **Refresh rate**: Each dashboard refresh queries Stripe API. Set appropriate refresh intervals (1m+ recommended).

## Development

```bash
# Clone and install
git clone https://github.com/jfreels123/grafana-stripe-datasource.git
cd grafana-stripe-datasource/jfreels123-stripe-datasource
npm install

# Terminal 1: Frontend
npm run dev

# Terminal 2: Backend (macOS)
mage -v build:darwin
# Or for Linux/Docker on Apple Silicon:
mage -v build:linuxARM64

# Terminal 3: Grafana
docker compose up
```

Open http://localhost:3000 (admin/admin)

### Testing

```bash
npm run e2e         # E2E tests
go test ./pkg/...   # Backend tests
npm run lint        # Linting
```

## License

Apache 2.0
