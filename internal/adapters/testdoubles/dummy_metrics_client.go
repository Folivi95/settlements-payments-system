package testdoubles

import "context"

type DummyMetricsClient struct{}

func (d DummyMetricsClient) Histogram(context.Context, string, float64, []string) {
}

func (d DummyMetricsClient) Count(context.Context, string, int64, []string) {
}
