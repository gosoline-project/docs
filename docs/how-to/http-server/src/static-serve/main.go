package main

import (
	"context"
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gosoline-project/httpserver"
	"github.com/justtrackio/gosoline/pkg/cfg"
	"github.com/justtrackio/gosoline/pkg/log"
)

//go:embed public
var publicFs embed.FS

func main() {
	httpserver.RunDefaultServer(func(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {
		router.GET("/api/status", func(ginCtx *gin.Context) {
			ginCtx.JSON(200, gin.H{"status": "ok"})
		})

		router.UseFactory(httpserver.CreateEmbeddedStaticServe(publicFs, "public", "/api"))

		return nil
	})
}

var _ = fs.FS(nil)
var _ = http.StatusOK
