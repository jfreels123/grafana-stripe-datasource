package stripe

import (
	"context"
	"time"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/balance"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/subscription"
)

type Client struct {
	key string
}

func NewClient(apiKey string) *Client {
	return &Client{key: apiKey}
}

type Metrics struct {
	MRR               int64
	ARR               int64
	ActiveSubscribers int64
	TotalCustomers    int64
	AvailableBalance  int64
	PendingBalance    int64
}

type SubscriptionData struct {
	ID        string
	Status    string
	Customer  string
	MRR       int64
	Created   time.Time
	PlanName  string
	Interval  string
}

func (c *Client) GetMetrics(ctx context.Context) (*Metrics, error) {
	stripe.Key = c.key

	m := &Metrics{}

	// Get subscriptions for MRR calculation
	subs, err := c.listActiveSubscriptions(ctx)
	if err != nil {
		return nil, err
	}

	for _, s := range subs {
		m.MRR += calculateMRR(s)
		m.ActiveSubscribers++
	}
	m.ARR = m.MRR * 12

	// Get customer count
	customers, err := c.countCustomers(ctx)
	if err != nil {
		return nil, err
	}
	m.TotalCustomers = customers

	// Get balance
	bal, err := c.getBalance(ctx)
	if err != nil {
		return nil, err
	}
	m.AvailableBalance = bal.Available
	m.PendingBalance = bal.Pending

	return m, nil
}

func (c *Client) GetSubscriptions(ctx context.Context) ([]SubscriptionData, error) {
	stripe.Key = c.key

	subs, err := c.listActiveSubscriptions(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]SubscriptionData, 0, len(subs))
	for _, s := range subs {
		sd := SubscriptionData{
			ID:       s.ID,
			Status:   string(s.Status),
			Customer: s.Customer.ID,
			MRR:      calculateMRR(s),
			Created:  time.Unix(s.Created, 0),
		}
		if len(s.Items.Data) > 0 {
			item := s.Items.Data[0]
			if item.Price != nil {
				sd.Interval = string(item.Price.Recurring.Interval)
				if item.Price.Product != nil {
					sd.PlanName = item.Price.Product.Name
				}
			}
		}
		result = append(result, sd)
	}
	return result, nil
}

func (c *Client) listActiveSubscriptions(ctx context.Context) ([]*stripe.Subscription, error) {
	params := &stripe.SubscriptionListParams{
		Status: stripe.String("active"),
	}
	params.Expand = []*string{
		stripe.String("data.items.data.price.product"),
	}
	params.Context = ctx

	var subs []*stripe.Subscription
	iter := subscription.List(params)
	for iter.Next() {
		subs = append(subs, iter.Subscription())
	}
	return subs, iter.Err()
}

func (c *Client) countCustomers(ctx context.Context) (int64, error) {
	params := &stripe.CustomerListParams{}
	params.Context = ctx

	var count int64
	iter := customer.List(params)
	for iter.Next() {
		count++
	}
	return count, iter.Err()
}

type BalanceInfo struct {
	Available int64
	Pending   int64
}

func (c *Client) getBalance(ctx context.Context) (*BalanceInfo, error) {
	params := &stripe.BalanceParams{}
	params.Context = ctx

	bal, err := balance.Get(params)
	if err != nil {
		return nil, err
	}

	info := &BalanceInfo{}
	for _, a := range bal.Available {
		if a.Currency == "usd" {
			info.Available = a.Amount
		}
	}
	for _, p := range bal.Pending {
		if p.Currency == "usd" {
			info.Pending = p.Amount
		}
	}
	return info, nil
}

// calculateMRR normalizes subscription amounts to monthly
func calculateMRR(s *stripe.Subscription) int64 {
	if len(s.Items.Data) == 0 {
		return 0
	}

	var total int64
	for _, item := range s.Items.Data {
		if item.Price == nil || item.Price.Recurring == nil {
			continue
		}
		amount := item.Price.UnitAmount * item.Quantity

		switch item.Price.Recurring.Interval {
		case stripe.PriceRecurringIntervalYear:
			total += amount / 12
		case stripe.PriceRecurringIntervalMonth:
			total += amount
		case stripe.PriceRecurringIntervalWeek:
			total += amount * 4
		case stripe.PriceRecurringIntervalDay:
			total += amount * 30
		}
	}
	return total
}

// Ping tests the API connection
func (c *Client) Ping(ctx context.Context) error {
	stripe.Key = c.key
	params := &stripe.BalanceParams{}
	params.Context = ctx
	_, err := balance.Get(params)
	return err
}
