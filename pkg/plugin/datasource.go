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
