# Serve a frontend

If you're building a single-page application (SPA), you can embed the frontend build into your Go binary and serve it directly from the httpserver.

## CreateEmbeddedStaticServe[​](#createembeddedstaticserve "Direct link to CreateEmbeddedStaticServe")

Use `CreateEmbeddedStaticServe` with Go's `embed.FS` to serve static files:

```
//go:embed public

var publicFs embed.FS
```

```
router.UseFactory(httpserver.CreateEmbeddedStaticServe(publicFs, "public", "/api"))
```

The three arguments are:

| Argument      | Description                                                      |
| ------------- | ---------------------------------------------------------------- |
| `files`       | The `embed.FS` containing your static files                      |
| `dir`         | The subtree root within the `embed.FS` (e.g., `"public"`)        |
| `excludes...` | Path prefixes to skip (e.g., `"/api"` — API routes handle these) |

## How it works[​](#how-it-works "Direct link to How it works")

1. For each incoming request, the middleware checks if the path starts with any excluded prefix
2. If excluded, the request passes through to the next handler
3. If not excluded, it looks for a matching file in the embedded filesystem
4. If the file has no extension, it falls back to serving `index.html` (SPA routing)
5. If no file is found, it returns 404

This means your API routes under `/api/*` are handled by your handlers, and everything else falls through to the SPA.

## Typical setup[​](#typical-setup "Direct link to Typical setup")

The common pattern is:

1. API routes handle `/api/*` paths
2. The static serve middleware handles everything else
3. The middleware is registered **after** API routes so API paths take priority

```
httpserver.RunDefaultServer(func(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {

    // API routes

    router.GET("/api/status", func(ginCtx *gin.Context) {

        ginCtx.JSON(200, gin.H{"status": "ok"})

    })



    // Static file serving (falls back to index.html for SPA routing)

    router.UseFactory(httpserver.CreateEmbeddedStaticServe(publicFs, "public", "/api"))



    return nil

})
```

## Directory structure[​](#directory-structure "Direct link to Directory structure")

Your project should look like this:

```
backend/

├── main.go              # //go:embed public

├── config.dist.yml

└── public/

    ├── index.html

    ├── assets/

    │   ├── index.js

    │   └── index.css

    └── favicon.ico
```

The `public/` directory is populated by your frontend build (e.g., `npm run build`).

## Complete example[​](#complete-example "Direct link to Complete example")

<!-- -->

main.go

main.go

```
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
```

config.dist.yml

config.dist.yml

```
app:

  env: dev

  name: static-serve



httpserver:

  default:

    port: 8088
```

Test:

```
# API endpoint

curl http://localhost:8088/api/status

# {"status":"ok"}



# Frontend

curl http://localhost:8088/

# <!DOCTYPE html>

# <html>...
```
