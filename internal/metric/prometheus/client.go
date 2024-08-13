package prometheus

import (
	api "github.com/prometheus/client_golang/api"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"

	rrethyv1 "github.com/RRethy/horizontalreplicascaler/api/v1"
)

type Option func(*Client)

type Client struct {
	DefaultApi prometheusv1.API
}

func NewClient() *Client {
	promClient, err := api.NewClient(api.Config{
		Address: "http://observe-prometheus-proxy.autoscaler-operator.svc.cluster.local:9090",
	})
	if err != nil {
		panic(err)
	}
	return &Client{
		DefaultApi: prometheusv1.NewAPI(promClient),
	}
}

func (c *Client) GetValue(metric rrethyv1.MetricSpec) (float64, error) {
	return 0, nil
}
