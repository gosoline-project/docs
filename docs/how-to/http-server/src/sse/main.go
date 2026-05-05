package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gosoline-project/httpserver"
	"github.com/justtrackio/gosoline/pkg/cfg"
	"github.com/justtrackio/gosoline/pkg/log"
)

func main() {
	httpserver.RunDefaultServer(func(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {
		router.Group("/api").HandleWith(httpserver.With(NewHandler, func(r *httpserver.Router, h *Handler) {
			r.GET("/events", httpserver.BindSseN(h.StreamEvents))
		}))

		return nil
	})
}

type Event struct {
	Message string `json:"message"`
	Time    string `json:"time"`
}

type Handler struct{}

func NewHandler(ctx context.Context, config cfg.Config, logger log.Logger) (*Handler, error) {
	return &Handler{}, nil
}

func (h *Handler) StreamEvents(ctx context.Context, writer *httpserver.SseWriter) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for i := 0; i < 5; i++ {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			data := Event{
				Message: fmt.Sprintf("event %d", i+1),
				Time:    time.Now().Format(time.RFC3339),
			}

			jsonData, err := json.Marshal(data)
			if err != nil {
				return fmt.Errorf("could not marshal event: %w", err)
			}

			err = writer.SendEvent(httpserver.SseEvent{
				Event: "update",
				Data:  string(jsonData),
				Id:    fmt.Sprintf("%d", i+1),
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
