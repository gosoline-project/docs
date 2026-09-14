# Stream with Server-Sent Events

Server-Sent Events (SSE) let you push real-time updates from the server to the client over a long-lived HTTP connection. The httpserver package has built-in support for SSE through the `SseWriter` type.

## Basic SSE handler[​](#basic-sse-handler "Direct link to Basic SSE handler")

Use `BindSseN` to create an SSE handler with no request input. The handler receives a `*SseWriter` instead of returning a `Response`:

```
r.GET("/events", httpserver.BindSseN(h.StreamEvents))
```

```
func (h *Handler) StreamEvents(ctx context.Context, writer *httpserver.SseWriter) error {

    ticker := time.NewTicker(time.Second)

    defer ticker.Stop()



    for i := 0; i < 5; i++ {

        select {

        case <-ctx.Done():

            return nil

        case <-ticker.C:

            err := writer.SendEvent(httpserver.SseEvent{

                Event: "update",

                Data:  fmt.Sprintf("event %d", i+1),

                Id:    fmt.Sprintf("%d", i+1),

            })

            if err != nil {

                return err

            }

        }

    }

    return nil

}
```

## Sending events[​](#sending-events "Direct link to Sending events")

`SseEvent` has three fields:

| Field   | Type   | Description                                                                                   |
| ------- | ------ | --------------------------------------------------------------------------------------------- |
| `Data`  | string | The event payload. Multi-line strings get split into multiple `data:` lines per the SSE spec. |
| `Event` | string | Optional event type name. Clients can listen for specific types using `addEventListener`.     |
| `Id`    | string | Optional event ID for reconnection support via `Last-Event-ID`.                               |
| `Retry` | int    | Optional reconnection hint in milliseconds.                                                   |

Use `writer.Send(data)` as a shortcut for `SendEvent(SseEvent{Data: data})`.

## SSE binding variants[​](#sse-binding-variants "Direct link to SSE binding variants")

| Function      | Signature                                    | Use case                       |
| ------------- | -------------------------------------------- | ------------------------------ |
| `BindSse[I]`  | `(ctx, *I, *SseWriter) error`                | SSE with request input binding |
| `BindSseR[I]` | `(ctx, *http.Request, *I, *SseWriter) error` | SSE with input + raw request   |
| `BindSseN`    | `(ctx, *SseWriter) error`                    | SSE with no input              |
| `BindSseNR`   | `(ctx, *http.Request, *SseWriter) error`     | SSE with raw request, no input |

## Exclude SSE paths from compression[​](#exclude-sse-paths-from-compression "Direct link to Exclude SSE paths from compression")

SSE requires streaming, so you must exclude SSE endpoints from gzip compression:

```
httpserver:

  default:

    port: 8088

    compression:

      exclude:

        path:

          - /api/events
```

Without this exclusion, gzip buffering will break the real-time stream.

## How it works[​](#how-it-works "Direct link to How it works")

* The `SseWriter` sets the correct headers: `Content-Type: text/event-stream`, `Cache-Control: no-cache`, `Connection: keep-alive`
* A heartbeat (`: heartbeat`) is sent every 5 seconds to keep connections alive
* When the client disconnects, `SendEvent` returns `ErrClientDisconnected`
* Errors in SSE handlers are sent as `event: error` SSE events, not as JSON error responses (which would corrupt the stream)

## Complete example[​](#complete-example "Direct link to Complete example")

<!-- -->

main.go

main.go

```
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
```

config.dist.yml

config.dist.yml

```
app:

  env: dev

  name: sse



httpserver:

  default:

    port: 8088

    compression:

      exclude:

        path:

          - /api/events
```

Test with `curl` or any SSE client:

```
curl -N http://localhost:8088/api/events

# event: update

# data: {"message":"event 1","time":"2026-05-05T10:00:00Z"}

# id: 1

#

# event: update

# data: {"message":"event 2","time":"2026-05-05T10:00:01Z"}

# id: 2
```
