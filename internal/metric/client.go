package metric

import (
	"fmt"

	rrethyv1 "github.com/RRethy/horizontalreplicascaler/api/v1"
	"github.com/RRethy/horizontalreplicascaler/internal/metric/prometheus"
	"github.com/RRethy/horizontalreplicascaler/internal/metric/static"
)

var _ Interface = &Client{}

type Interface interface {
	GetValue(rrethyv1.MetricSpec) (float64, error)
}

type Option func(*Client)

func WithStaticClient(client Interface) Option {
	return func(c *Client) { c.staticClient = client }
}

func WithPrometheusClient(client Interface) Option {
	return func(c *Client) { c.prometheusClient = client }
}

type Client struct {
	staticClient     Interface
	prometheusClient Interface
}

func NewClient(opts ...Option) *Client {
	client := &Client{
		staticClient:     static.NewClient(),
		prometheusClient: prometheus.NewClient(),
	}
	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (c *Client) GetValue(metric rrethyv1.MetricSpec) (float64, error) {
	switch metric.Type {
	case rrethyv1.StaticMetricType:
		return c.staticClient.GetValue(metric)
	case rrethyv1.PrometheusMetricType:
		return c.prometheusClient.GetValue(metric)
	default:
		return 0, fmt.Errorf("unknown metric type %s", metric.Type)
	}
}
