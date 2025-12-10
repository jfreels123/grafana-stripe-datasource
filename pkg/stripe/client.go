package stripe

import (
	"context"
	"time"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/balance"
	"github.com/stripe/stripe-go/v82/charge"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/invoice"
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
	// SaaS metrics matching Stripe dashboard
	NewMRR           int64   // MRR from subs created in last 30 days
	ChurnedMRR       int64   // MRR lost from canceled subs in last 30 days
	NetNewMRR        int64   // New MRR - Churned MRR
	ChurnRate        float64 // Churned / (Active 30 days ago + New)
	ARPU             int64   // MRR / ActiveSubscribers
	TrialingCount    int64   // Subscriptions currently in trial
	PastDueCount     int64   // Subscriptions past due
	CanceledCount30d int64   // Canceled in last 30 days
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
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30).Unix()

	// Get active subscriptions for MRR calculation
	subs, err := c.listActiveSubscriptions(ctx)
	if err != nil {
		return nil, err
	}

	for _, s := range subs {
		mrr := calculateMRR(s)
		m.MRR += mrr
		m.ActiveSubscribers++

		// New MRR = subscriptions created in last 30 days
		if s.Created >= thirtyDaysAgo {
			m.NewMRR += mrr
		}
	}
	m.ARR = m.MRR * 12

	// Calculate ARPU
	if m.ActiveSubscribers > 0 {
		m.ARPU = m.MRR / m.ActiveSubscribers
	}

	// Get trialing subscriptions
	trialing, err := c.countSubscriptionsByStatus(ctx, "trialing")
	if err != nil {
		return nil, err
	}
	m.TrialingCount = trialing

	// Get past due subscriptions
	pastDue, err := c.countSubscriptionsByStatus(ctx, "past_due")
	if err != nil {
		return nil, err
	}
	m.PastDueCount = pastDue

	// Get canceled subscriptions in last 30 days for churn
	canceled, churnedMRR, err := c.getCanceledSubscriptions(ctx, thirtyDaysAgo)
	if err != nil {
		return nil, err
	}
	m.CanceledCount30d = canceled
	m.ChurnedMRR = churnedMRR
	m.NetNewMRR = m.NewMRR - m.ChurnedMRR

	// Calculate churn rate: churned / (active 30 days ago + new)
	// Approximate: active 30 days ago ≈ current active - new + churned
	active30DaysAgo := m.ActiveSubscribers - (m.NewMRR / max(m.ARPU, 1)) + m.CanceledCount30d
	if active30DaysAgo > 0 {
		m.ChurnRate = float64(m.CanceledCount30d) / float64(active30DaysAgo) * 100
	}

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

func (c *Client) countSubscriptionsByStatus(ctx context.Context, status string) (int64, error) {
	params := &stripe.SubscriptionListParams{
		Status: stripe.String(status),
	}
	params.Context = ctx

	var count int64
	iter := subscription.List(params)
	for iter.Next() {
		count++
	}
	return count, iter.Err()
}

func (c *Client) getCanceledSubscriptions(ctx context.Context, since int64) (int64, int64, error) {
	params := &stripe.SubscriptionListParams{
		Status: stripe.String("canceled"),
	}
	params.Expand = []*string{
		stripe.String("data.items.data.price"),
	}
	params.Context = ctx

	var count int64
	var churnedMRR int64
	iter := subscription.List(params)
	for iter.Next() {
		s := iter.Subscription()
		// Only count if canceled in the time window
		if s.CanceledAt >= since {
			count++
			churnedMRR += calculateMRR(s)
		}
	}
	return count, churnedMRR, iter.Err()
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
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
				if item.Price.Recurring != nil {
					sd.Interval = string(item.Price.Recurring.Interval)
				}
				sd.PlanName = item.Price.Nickname
				if sd.PlanName == "" && item.Price.Product != nil {
					sd.PlanName = item.Price.Product.ID
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
	// Only expand to 4 levels (Stripe's limit)
	params.Expand = []*string{
		stripe.String("data.items.data.price"),
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

// InvoiceData represents invoice information
type InvoiceData struct {
	ID           string
	Customer     string
	Status       string
	Amount       int64
	AmountPaid   int64
	Currency     string
	Created      time.Time
	DueDate      time.Time
	Paid         bool
	ProductName  string
}

// GetInvoices returns recent invoices
func (c *Client) GetInvoices(ctx context.Context) ([]InvoiceData, error) {
	stripe.Key = c.key

	params := &stripe.InvoiceListParams{}
	params.Limit = stripe.Int64(100)
	params.Context = ctx

	var invoices []InvoiceData
	iter := invoice.List(params)
	for iter.Next() {
		inv := iter.Invoice()
		isPaid := inv.Status == stripe.InvoiceStatusPaid
		data := InvoiceData{
			ID:         inv.ID,
			Status:     string(inv.Status),
			Amount:     inv.Total,
			AmountPaid: inv.AmountPaid,
			Currency:   string(inv.Currency),
			Created:    time.Unix(inv.Created, 0),
			Paid:       isPaid,
		}
		if inv.Customer != nil {
			data.Customer = inv.Customer.ID
		}
		if inv.DueDate > 0 {
			data.DueDate = time.Unix(inv.DueDate, 0)
		}
		invoices = append(invoices, data)
	}
	return invoices, iter.Err()
}

// InvoiceMetrics represents aggregated invoice metrics
type InvoiceMetrics struct {
	TotalRevenue    int64
	PaidInvoices    int64
	UnpaidInvoices  int64
	OverdueInvoices int64
}

// GetInvoiceMetrics returns aggregated invoice metrics
func (c *Client) GetInvoiceMetrics(ctx context.Context) (*InvoiceMetrics, error) {
	stripe.Key = c.key

	params := &stripe.InvoiceListParams{}
	params.Context = ctx

	m := &InvoiceMetrics{}
	iter := invoice.List(params)
	for iter.Next() {
		inv := iter.Invoice()
		if inv.Status == stripe.InvoiceStatusPaid {
			m.TotalRevenue += inv.AmountPaid
			m.PaidInvoices++
		} else {
			m.UnpaidInvoices++
			if inv.DueDate > 0 && time.Unix(inv.DueDate, 0).Before(time.Now()) {
				m.OverdueInvoices++
			}
		}
	}
	return m, iter.Err()
}

// ChargeData represents charge information
type ChargeData struct {
	ID       string
	Amount   int64
	Currency string
	Status   string
	Customer string
	Created  time.Time
	Paid     bool
	Refunded bool
}

// GetCharges returns recent charges
func (c *Client) GetCharges(ctx context.Context) ([]ChargeData, error) {
	stripe.Key = c.key

	params := &stripe.ChargeListParams{}
	params.Limit = stripe.Int64(100)
	params.Context = ctx

	var charges []ChargeData
	iter := charge.List(params)
	for iter.Next() {
		ch := iter.Charge()
		data := ChargeData{
			ID:       ch.ID,
			Amount:   ch.Amount,
			Currency: string(ch.Currency),
			Status:   string(ch.Status),
			Created:  time.Unix(ch.Created, 0),
			Paid:     ch.Paid,
			Refunded: ch.Refunded,
		}
		if ch.Customer != nil {
			data.Customer = ch.Customer.ID
		}
		charges = append(charges, data)
	}
	return charges, iter.Err()
}

// ChargeMetrics represents aggregated charge metrics
type ChargeMetrics struct {
	TotalCharges     int64
	SuccessfulAmount int64
	FailedCount      int64
	RefundedCount    int64
	RefundedAmount   int64
}

// GetChargeMetrics returns aggregated charge metrics
func (c *Client) GetChargeMetrics(ctx context.Context) (*ChargeMetrics, error) {
	stripe.Key = c.key

	params := &stripe.ChargeListParams{}
	params.Context = ctx

	m := &ChargeMetrics{}
	iter := charge.List(params)
	for iter.Next() {
		ch := iter.Charge()
		m.TotalCharges++
		if ch.Paid {
			m.SuccessfulAmount += ch.Amount
		}
		if ch.Status == stripe.ChargeStatusFailed {
			m.FailedCount++
		}
		if ch.Refunded {
			m.RefundedCount++
			m.RefundedAmount += ch.AmountRefunded
		}
	}
	return m, iter.Err()
}

// ProductRevenue represents revenue by product
type ProductRevenue struct {
	ProductID   string
	ProductName string
	Revenue     int64
	SubCount    int64
}

// GetRevenueByProduct returns revenue breakdown by product
func (c *Client) GetRevenueByProduct(ctx context.Context) ([]ProductRevenue, error) {
	stripe.Key = c.key

	subs, err := c.listActiveSubscriptions(ctx)
	if err != nil {
		return nil, err
	}

	productMap := make(map[string]*ProductRevenue)
	for _, s := range subs {
		for _, item := range s.Items.Data {
			if item.Price == nil {
				continue
			}
			// Use price nickname or product ID as identifier
			pid := ""
			pname := item.Price.Nickname
			if item.Price.Product != nil {
				pid = item.Price.Product.ID
				if pname == "" {
					pname = item.Price.Product.ID
				}
			}
			if pid == "" {
				pid = item.Price.ID
			}
			if pname == "" {
				pname = item.Price.ID
			}
			mrr := calculateMRR(s)

			if pr, ok := productMap[pid]; ok {
				pr.Revenue += mrr
				pr.SubCount++
			} else {
				productMap[pid] = &ProductRevenue{
					ProductID:   pid,
					ProductName: pname,
					Revenue:     mrr,
					SubCount:    1,
				}
			}
		}
	}

	result := make([]ProductRevenue, 0, len(productMap))
	for _, pr := range productMap {
		result = append(result, *pr)
	}
	return result, nil
}
