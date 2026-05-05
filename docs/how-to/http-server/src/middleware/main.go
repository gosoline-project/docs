package main

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gosoline-project/httpserver"
	"github.com/justtrackio/gosoline/pkg/cfg"
	"github.com/justtrackio/gosoline/pkg/log"
)

func main() {
	httpserver.RunDefaultServer(func(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {
		router.Use(func(ginCtx *gin.Context) {
			start := time.Now()
			ginCtx.Next()
			elapsed := time.Since(start)
			logger.Info(ctx, "request completed in %s", elapsed)
		})

		router.UseFactory(func(ctx context.Context, config cfg.Config, logger log.Logger) (gin.HandlerFunc, error) {
			return func(ginCtx *gin.Context) {
				ginCtx.Header("X-Server-Name", "my-server")
				ginCtx.Next()
			}, nil
		})

		router.GET("/ping", func(ginCtx *gin.Context) {
			ginCtx.JSON(200, gin.H{"message": "pong"})
		})

		return nil
	})
}
