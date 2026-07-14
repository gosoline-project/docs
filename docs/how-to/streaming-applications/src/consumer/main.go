package main

import (
	"context"

	"github.com/justtrackio/gosoline/pkg/application"
	"github.com/justtrackio/gosoline/pkg/cfg"
	"github.com/justtrackio/gosoline/pkg/log"
	"github.com/justtrackio/gosoline/pkg/stream"
)

type OrderCreated struct {
	Id    string  `json:"id"`
	Total float64 `json:"total"`
}

type consumer struct {
	logger log.Logger
}

func main() {
	application.RunConsumer(newConsumer,
		application.WithConfigFile("config.dist.yml", "yml"),
	)
}

func newConsumer(_ context.Context, _ cfg.Config, logger log.Logger) (stream.ConsumerCallback[OrderCreated], error) {
	return &consumer{logger: logger}, nil
}

func (c *consumer) Consume(ctx context.Context, order OrderCreated, _ map[string]string) (bool, error) {
	c.logger.Info(ctx, "received order %s with total %.2f", order.Id, order.Total)

	return true, nil
}
