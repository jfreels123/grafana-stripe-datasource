package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/jfreels123/stripe-datasource/pkg/models"
	"github.com/jfreels123/stripe-datasource/pkg/stripe"
)

var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

type Datasource struct {
	client *stripe.Client
}

func NewDatasource(_ context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	config, err := models.LoadPluginSettings(settings)
	if err != nil {
		return nil, err
	}
	return &Datasource{
		client: stripe.NewClient(config.Secrets.ApiKey),
	}, nil
}

func (d *Datasource) Dispose() {}

func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()
	for _, q := range req.Queries {
		response.Responses[q.RefID] = d.query(ctx, q)
	}
	return response, nil
}

type QueryType string

const (
	QueryMRR           QueryType = "mrr"
	QueryARR           QueryType = "arr"
	QuerySubscribers   QueryType = "subscribers"
	QueryCustomers     QueryType = "customers"
	QueryBalance       QueryType = "balance"
	QuerySubscriptions QueryType = "subscriptions"
	QueryRevenue       QueryType = "revenue"
	QueryInvoices      QueryType = "invoices"
	QueryCharges       QueryType = "charges"
	QueryProducts      QueryType = "products"
	// New Stripe dashboard metrics
	QueryNewMRR       QueryType = "new_mrr"
	QueryChurnedMRR   QueryType = "churned_mrr"
	QueryNetNewMRR    QueryType = "net_new_mrr"
	QueryChurnRate    QueryType = "churn_rate"
	QueryARPU         QueryType = "arpu"
	QueryTrialing     QueryType = "trialing"
	QueryPastDue      QueryType = "past_due"
)

type queryModel struct {
	QueryType QueryType `json:"queryType"`
}

func (d *Datasource) query(ctx context.Context, q backend.DataQuery) backend.DataResponse {
	var qm queryModel
	if err := json.Unmarshal(q.JSON, &qm); err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err))
	}

	switch qm.QueryType {
	case QuerySubscriptions:
		return d.querySubscriptions(ctx, q)
	case QueryInvoices:
		return d.queryInvoices(ctx, q)
	case QueryCharges:
		return d.queryCharges(ctx, q)
	case QueryProducts:
		return d.queryProducts(ctx, q)
	case QueryRevenue:
		return d.queryRevenue(ctx, q)
	default:
		return d.queryMetrics(ctx, q, qm.QueryType)
	}
}

func (d *Datasource) queryMetrics(ctx context.Context, q backend.DataQuery, queryType QueryType) backend.DataResponse {
	metrics, err := d.client.GetMetrics(ctx)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("stripe error: %v", err))
	}

	now := time.Now()
	frame := data.NewFrame("metrics")
	frame.Meta = &data.FrameMeta{
		PreferredVisualizationPluginID: "stat",
	}

	var value float64
	var name string

	switch queryType {
	case QueryMRR:
		value = float64(metrics.MRR) / 100
		name = "MRR"
	case QueryARR:
		value = float64(metrics.ARR) / 100
		name = "ARR"
	case QuerySubscribers:
		value = float64(metrics.ActiveSubscribers)
		name = "Active Subscribers"
	case QueryCustomers:
		value = float64(metrics.TotalCustomers)
		name = "Total Customers"
	case QueryBalance:
		value = float64(metrics.AvailableBalance) / 100
		name = "Available Balance"
	case QueryNewMRR:
		value = float64(metrics.NewMRR) / 100
		name = "New MRR"
	case QueryChurnedMRR:
		value = float64(metrics.ChurnedMRR) / 100
		name = "Churned MRR"
	case QueryNetNewMRR:
		value = float64(metrics.NetNewMRR) / 100
		name = "Net New MRR"
	case QueryChurnRate:
		value = metrics.ChurnRate
		name = "Churn Rate %"
	case QueryARPU:
		value = float64(metrics.ARPU) / 100
		name = "ARPU"
	case QueryTrialing:
		value = float64(metrics.TrialingCount)
		name = "Trialing"
	case QueryPastDue:
		value = float64(metrics.PastDueCount)
		name = "Past Due"
	default:
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("unknown query type: %s", queryType))
	}

	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, []time.Time{now}),
		data.NewField(name, nil, []float64{value}),
	)

	return backend.DataResponse{Frames: []*data.Frame{frame}}
}

func (d *Datasource) querySubscriptions(ctx context.Context, q backend.DataQuery) backend.DataResponse {
	subs, err := d.client.GetSubscriptions(ctx)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("stripe error: %v", err))
	}

	frame := data.NewFrame("subscriptions")
	frame.Meta = &data.FrameMeta{
		PreferredVisualization: data.VisTypeTable,
	}

	ids := make([]string, len(subs))
	statuses := make([]string, len(subs))
	customers := make([]string, len(subs))
	mrrs := make([]float64, len(subs))
	plans := make([]string, len(subs))
	intervals := make([]string, len(subs))
	created := make([]time.Time, len(subs))

	for i, s := range subs {
		ids[i] = s.ID
		statuses[i] = s.Status
		customers[i] = s.Customer
		mrrs[i] = float64(s.MRR) / 100
		plans[i] = s.PlanName
		intervals[i] = s.Interval
		created[i] = s.Created
	}

	frame.Fields = append(frame.Fields,
		data.NewField("id", nil, ids),
		data.NewField("status", nil, statuses),
		data.NewField("customer", nil, customers),
		data.NewField("mrr", nil, mrrs),
		data.NewField("plan", nil, plans),
		data.NewField("interval", nil, intervals),
		data.NewField("created", nil, created),
	)

	return backend.DataResponse{Frames: []*data.Frame{frame}}
}

func (d *Datasource) queryInvoices(ctx context.Context, q backend.DataQuery) backend.DataResponse {
	invoices, err := d.client.GetInvoices(ctx)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("stripe error: %v", err))
	}

	frame := data.NewFrame("invoices")
	frame.Meta = &data.FrameMeta{
		PreferredVisualization: data.VisTypeTable,
	}

	ids := make([]string, len(invoices))
	customers := make([]string, len(invoices))
	statuses := make([]string, len(invoices))
	amounts := make([]float64, len(invoices))
	amountsPaid := make([]float64, len(invoices))
	created := make([]time.Time, len(invoices))
	paid := make([]bool, len(invoices))

	for i, inv := range invoices {
		ids[i] = inv.ID
		customers[i] = inv.Customer
		statuses[i] = inv.Status
		amounts[i] = float64(inv.Amount) / 100
		amountsPaid[i] = float64(inv.AmountPaid) / 100
		created[i] = inv.Created
		paid[i] = inv.Paid
	}

	frame.Fields = append(frame.Fields,
		data.NewField("id", nil, ids),
		data.NewField("customer", nil, customers),
		data.NewField("status", nil, statuses),
		data.NewField("amount", nil, amounts),
		data.NewField("amount_paid", nil, amountsPaid),
		data.NewField("created", nil, created),
		data.NewField("paid", nil, paid),
	)

	return backend.DataResponse{Frames: []*data.Frame{frame}}
}

func (d *Datasource) queryCharges(ctx context.Context, q backend.DataQuery) backend.DataResponse {
	charges, err := d.client.GetCharges(ctx)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("stripe error: %v", err))
	}

	frame := data.NewFrame("charges")
	frame.Meta = &data.FrameMeta{
		PreferredVisualization: data.VisTypeTable,
	}

	ids := make([]string, len(charges))
	customers := make([]string, len(charges))
	statuses := make([]string, len(charges))
	amounts := make([]float64, len(charges))
	created := make([]time.Time, len(charges))
	paid := make([]bool, len(charges))
	refunded := make([]bool, len(charges))

	for i, ch := range charges {
		ids[i] = ch.ID
		customers[i] = ch.Customer
		statuses[i] = ch.Status
		amounts[i] = float64(ch.Amount) / 100
		created[i] = ch.Created
		paid[i] = ch.Paid
		refunded[i] = ch.Refunded
	}

	frame.Fields = append(frame.Fields,
		data.NewField("id", nil, ids),
		data.NewField("customer", nil, customers),
		data.NewField("status", nil, statuses),
		data.NewField("amount", nil, amounts),
		data.NewField("created", nil, created),
		data.NewField("paid", nil, paid),
		data.NewField("refunded", nil, refunded),
	)

	return backend.DataResponse{Frames: []*data.Frame{frame}}
}

func (d *Datasource) queryProducts(ctx context.Context, q backend.DataQuery) backend.DataResponse {
	products, err := d.client.GetRevenueByProduct(ctx)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("stripe error: %v", err))
	}

	frame := data.NewFrame("products")
	frame.Meta = &data.FrameMeta{
		PreferredVisualization: data.VisTypeTable,
	}

	names := make([]string, len(products))
	revenues := make([]float64, len(products))
	subCounts := make([]int64, len(products))

	for i, p := range products {
		names[i] = p.ProductName
		revenues[i] = float64(p.Revenue) / 100
		subCounts[i] = p.SubCount
	}

	frame.Fields = append(frame.Fields,
		data.NewField("product", nil, names),
		data.NewField("mrr", nil, revenues),
		data.NewField("subscriptions", nil, subCounts),
	)

	return backend.DataResponse{Frames: []*data.Frame{frame}}
}

func (d *Datasource) queryRevenue(ctx context.Context, q backend.DataQuery) backend.DataResponse {
	metrics, err := d.client.GetInvoiceMetrics(ctx)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("stripe error: %v", err))
	}

	now := time.Now()
	frame := data.NewFrame("revenue")
	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, []time.Time{now}),
		data.NewField("Total Revenue", nil, []float64{float64(metrics.TotalRevenue) / 100}),
	)

	return backend.DataResponse{Frames: []*data.Frame{frame}}
}

func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	config, err := models.LoadPluginSettings(*req.PluginContext.DataSourceInstanceSettings)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Unable to load settings",
		}, nil
	}

	if config.Secrets.ApiKey == "" {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "API key is missing",
		}, nil
	}

	client := stripe.NewClient(config.Secrets.ApiKey)
	if err := client.Ping(ctx); err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Stripe API error: %v", err),
		}, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Connected to Stripe",
	}, nil
}
