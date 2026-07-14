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
	Id string `json:"id"`
}

type producerModule struct {
	producer stream.Producer
}

func main() {
	application.RunModule("producer", newModule,
		application.WithConfigFile("config.dist.yml", "yml"),
		application.WithProducerDaemon,
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
	for i := 1; i <= 5; i++ {
		order := OrderCreated{Id: fmt.Sprintf("order-%04d", i)}
		if err := m.producer.WriteOne(ctx, order); err != nil {
			return fmt.Errorf("publish order: %w", err)
		}
	}

	return nil
}
