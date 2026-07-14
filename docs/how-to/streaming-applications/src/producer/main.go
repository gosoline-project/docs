package main

import (
	"context"
	"fmt"

	"github.com/justtrackio/gosoline/pkg/application"
	"github.com/justtrackio/gosoline/pkg/cfg"
	"github.com/justtrackio/gosoline/pkg/kernel"
	"github.com/justtrackio/gosoline/pkg/log"
	"github.com/justtrackio/gosoline/pkg/stream"
)

type OrderCreated struct {
	Id    string  `json:"id"`
	Total float64 `json:"total"`
}

type producerModule struct {
	producer stream.Producer
}

func main() {
	application.RunModule("producer", newModule,
		application.WithConfigFile("config.dist.yml", "yml"),
	)
}

func newModule(ctx context.Context, config cfg.Config, logger log.Logger) (kernel.Module, error) {
	producer, err := stream.NewProducer(ctx, config, logger, "orders")
	if err != nil {
		return nil, fmt.Errorf("create orders producer: %w", err)
	}

	return &producerModule{producer: producer}, nil
}

func (m *producerModule) Run(ctx context.Context) error {
	if err := m.producer.WriteOne(ctx, OrderCreated{
		Id:    "order-1001",
		Total: 42.50,
	}); err != nil {
		return fmt.Errorf("publish order: %w", err)
	}

	orders := []OrderCreated{
		{Id: "order-1002", Total: 18.75},
		{Id: "order-1003", Total: 91.20},
	}

	if err := m.producer.Write(ctx, orders); err != nil {
		return fmt.Errorf("publish orders: %w", err)
	}

	return nil
}
