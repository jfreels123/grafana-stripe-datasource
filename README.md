# Stripe Data Source for Grafana

A Grafana data source plugin for visualizing Stripe metrics including MRR, ARR, subscriber counts, and more.

## Features

- **MRR** - Monthly Recurring Revenue calculated from active subscriptions
- **ARR** - Annual Recurring Revenue (MRR x 12)
- **Active Subscribers** - Count of active subscriptions
- **Total Customers** - Customer count
- **Available Balance** - USD balance available for payout
- **Subscriptions** - Table of all active subscriptions with details

## Requirements

- Grafana 10.4.0+
- Stripe API key (read-only scope recommended)

## Installation

1. Download the latest release
2. Extract to your Grafana plugins directory
3. Restart Grafana

## Configuration

1. Go to **Connections > Data sources > Add data source**
2. Search for "Stripe"
3. Enter your Stripe API key (`sk_live_...` or `sk_test_...`)
4. Click **Save & test**

### API Key Permissions

Use a restricted API key with read-only access to:
- Customers (read)
- Subscriptions (read)
- Balance (read)

## Usage

### Single Value Metrics

Use **Stat** or **Gauge** visualization and select a metric (MRR, ARR, etc.)

### Subscriptions Table

Use **Table** visualization with "Subscriptions" metric to see all active subscriptions.

## Development

```bash
# Install dependencies
npm install

# Start frontend dev server
npm run dev

# Build backend (requires mage)
mage -v build:darwin  # or build:linux

# Start Grafana dev environment
docker compose up

# Open http://localhost:3000 (admin/admin)
```

## Distributing

See the [Grafana plugin publishing docs](https://grafana.com/developers/plugin-tools/publish-a-plugin/sign-a-plugin).

## License

Apache 2.0
