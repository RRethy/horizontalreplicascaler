package static

import (
	"fmt"
	"strconv"

	rrethyv1 "github.com/RRethy/horizontalreplicascaler/api/v1"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) GetValue(metric rrethyv1.MetricSpec) (float64, error) {
	target, err := strconv.ParseFloat(metric.Target.Value, 64)
	if err != nil {
		return 0, fmt.Errorf("failed parsing target value %s: %w", metric.Target, err)
	}
	return target, nil
}
