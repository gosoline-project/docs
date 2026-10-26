# Implement health checks

You can check the health of a running gosoline application.

In this guide, you'll learn how to:

* Configure and use the on-run health check.
* Configure and use the health-check endpoint.
* Customize a module's health-check behavior.

## Check your application on run[​](#check-your-application-on-run "Direct link to Check your application on run")

When you run your application, the kernel automatically performs a health check. After you call run, the application transitions into the "running" state only after all modules are considered healthy. You can configure the timing behavior of this health check like this:

```
kernel:

  health_check:    

    timeout: 10s    

    wait_interval: 1s
```

Here, you:

1. Configure an on-run health check.
2. Establish a timeout for the health check.
3. Establish a one-second wait interval before checking or re-checking if the app is healthy.

The `kernel.health_check.wait_interval` is used to verify that all modules are healthy. If they are not all healthy by the time the `kernel.health_check.timeout` is reached, the application stops. Once the application is considered healthy on run, the kernel no longer performs its own health checks.

Therefore, to check the health of your application once it's running, you need to use the health-check endpoint.

## Check your app's health manually[​](#check-your-apps-health-manually "Direct link to Check your app's health manually")

To use the health-check endpoint, you need to explicitly enable it by passing `application.WithHttpHealthCheck` as an option when creating your application.

Once your application is running, you can perform a health check with a simple HTTP request to its health-check endpoint:

```
GET /health HTTP/1.1

Host: localhost:8090
```

You can configure the health-check path and port under the `httpserver.health-check` key:

```
httpserver:

  health-check:

    path: /health

    port: 8090
```

By default, this route responds with a `200 OK` status code if the application is considered healthy.

If you want to have more control over the module's health status, you can implement the `kernel.HealthCheckedModule` interface for your application modules.

## Customize your health check[​](#customize-your-health-check "Direct link to Customize your health check")

To control how the health of your application is verified, you'll implement the `kernel.HealthCheckedModule` interface:

```
package kernel



type HealthCheckedModule interface {

	IsHealthy(ctx context.Context) (bool, error)

}
```

Here you'll walk through a complete example application of how you might implement this interface, starting with `module.go`.

### Implement module.go[​](#implement-modulego "Direct link to Implement module.go")

In `module.go`, add the following code:

module.go

```
package main



import (

	"context"

	"sync/atomic"

	"time"



	"github.com/justtrackio/gosoline/pkg/cfg"

	"github.com/justtrackio/gosoline/pkg/clock"

	"github.com/justtrackio/gosoline/pkg/kernel"

	"github.com/justtrackio/gosoline/pkg/log"

)



func NewHelloWorldModule(ctx context.Context, config cfg.Config, logger log.Logger) (kernel.Module, error) {

	return &HelloWorldModule{

		logger: logger.WithChannel("hello-world"),

	}, nil

}



type HelloWorldModule struct {

	logger  log.Logger

	healthy atomic.Bool

}



func (h *HelloWorldModule) IsHealthy(ctx context.Context) (bool, error) {

	return h.healthy.Load(), nil

}



func (h *HelloWorldModule) Run(ctx context.Context) error {

	timer := clock.NewRealTimer(time.Second * 3)

	<-timer.Chan()



	h.healthy.Store(true)



	h.logger.Info(ctx, "Hello World")



	return nil

}
```

Now, you'll walkthrough this file in detail to learn how it works.

First, you declare `healthy`, a boolean to store the health of the module:

```
type HelloWorldModule struct {

	logger log.Logger

	healthy atomic.Bool

}
```

Then, you implement `IsHealthy()`, a function that gets called on every health check:

```
func (h *HelloWorldModule) IsHealthy(ctx context.Context) (bool, error) {

	return h.healthy.Load(), nil

}
```

Next, you add a timer that simulates some work that has to be done before the module is considered healthy, then set `healthy` to `true`:

```
func (h *HelloWorldModule) Run(ctx context.Context) error {

	timer := clock.NewRealTimer(time.Second * 3)

	<-timer.Chan()



	h.healthy.Store(true)



	h.logger.Info(ctx, "Hello World")



	return nil

}
```

And that's it! You've implemented custom health-check logic in your module.

Now, add your app's `main.go` file:

main.go

```
package main



import "github.com/justtrackio/gosoline/pkg/application"



func main() {

	application.RunModule("hello-world", NewHelloWorldModule,

		application.WithConfigFile("config.dist.yml", "yml"),

		application.WithHttpHealthCheck,

	)

}
```

Then, add your app's configuration, including health-checks:

config.dist.yml.go

```
app:

  env: dev

  name: hello-world



httpserver:

  health-check:

    path: /health

    port: 8090



kernel:

  health_check:

    timeout: 10s

    wait_interval: 1s
```

And when you run your app, you'll see something like the following output:

stdout

```
15:09:29.102 main    info    applied priority 1 config post processor 'gosoline.dx.useRandomPort'  application: hello-world

15:09:29.102 main    info    applied priority 1 config post processor 'gosoline.reslife.autoCreate'  application: hello-world

15:09:29.102 main    info    applied priority 8 config post processor 'gosoline.log.handler_main'  application: hello-world

15:09:29.103 kernel  info    starting kernel                                     application: hello-world

15:09:29.103 kernel  info    cfg app.env=dev                                     application: hello-world

15:09:29.103 kernel  info    cfg app.name=hello-world                            application: hello-world

15:09:29.103 kernel  info    cfg appctx.metadata.server.port=0                   application: hello-world

15:09:29.103 kernel  info    cfg dx.use_random_port=true                         application: hello-world

15:09:29.103 kernel  info    cfg httpserver.health-check.path=/health            application: hello-world

15:09:29.103 kernel  info    cfg httpserver.health-check.port=8090               application: hello-world

15:09:29.103 kernel  info    cfg httpserver.health-check.timeout.drain=0s        application: hello-world

15:09:29.103 kernel  info    cfg httpserver.health-check.timeout.idle=1m0s       application: hello-world

15:09:29.103 kernel  info    cfg httpserver.health-check.timeout.read=1m0s       application: hello-world

15:09:29.103 kernel  info    cfg httpserver.health-check.timeout.shutdown=1m0s   application: hello-world

15:09:29.103 kernel  info    cfg httpserver.health-check.timeout.write=1m0s      application: hello-world

15:09:29.103 kernel  info    cfg kernel.health_check.timeout=10s                 application: hello-world

15:09:29.103 kernel  info    cfg kernel.health_check.wait_interval=1s            application: hello-world

15:09:29.103 kernel  info    cfg kernel.kill_timeout=30s                         application: hello-world

15:09:29.103 kernel  info    cfg kinsumer.autoscale.dynamodb.naming.pattern={app.env}-kinsumer-autoscale-leaders  application: hello-world

15:09:29.103 kernel  info    cfg kinsumer.autoscale.ecs.client=default           application: hello-world

15:09:29.103 kernel  info    cfg kinsumer.autoscale.ecs.cluster={app.env}        application: hello-world

15:09:29.103 kernel  info    cfg kinsumer.autoscale.ecs.service={app.name}       application: hello-world

15:09:29.103 kernel  info    cfg kinsumer.autoscale.enabled=false                application: hello-world

15:09:29.103 kernel  info    cfg kinsumer.autoscale.leader_election=kinsumer-autoscale  application: hello-world

15:09:29.103 kernel  info    cfg kinsumer.autoscale.orchestrator=ecs             application: hello-world

15:09:29.103 kernel  info    cfg kinsumer.autoscale.period=1m0s                  application: hello-world

15:09:29.103 kernel  info    cfg log.handlers.main.channels.kernel.level=        application: hello-world

15:09:29.103 kernel  info    cfg log.handlers.main.channels.main.level=          application: hello-world

15:09:29.103 kernel  info    cfg log.handlers.main.formatter=console             application: hello-world

15:09:29.103 kernel  info    cfg log.handlers.main.level=info                    application: hello-world

15:09:29.103 kernel  info    cfg log.handlers.main.timestamp_format=15:04:05.000  application: hello-world

15:09:29.103 kernel  info    cfg log.handlers.main.type=iowriter                 application: hello-world

15:09:29.103 kernel  info    cfg log.handlers.main.writer=stdout                 application: hello-world

15:09:29.103 kernel  info    cfg resource_lifecycles.create.enabled=true         application: hello-world

15:09:29.103 kernel  info    cfg fingerprint: c3bbbde8d0193517f02ef31853922032   application: hello-world

15:09:29.103 kernel  info    stage 0 up and running with 1 modules               application: hello-world

15:09:29.103 kernel  info    waiting 1s for module hello-world in stage 2048 to get healthy: time left 9.9999975s  application: hello-world

15:09:29.103 kernel  info    running foreground module hello-world in stage 2048  application: hello-world

15:09:29.103 kernel  info    running background module http-health-check in stage 0  application: hello-world

15:09:30.104 kernel  info    waiting 1s for module hello-world in stage 2048 to get healthy: time left 8.998905459s  application: hello-world

15:09:31.104 kernel  info    waiting 1s for module hello-world in stage 2048 to get healthy: time left 7.998746834s  application: hello-world

15:09:32.103 hello-world info    Hello World                                         application: hello-world

15:09:32.103 kernel  info    stopped foreground module hello-world               application: hello-world

15:09:32.103 kernel  info    stopping kernel due to: no more foreground modules in running state  application: hello-world

15:09:32.103 kernel  info    stopping stage 2048                                 application: hello-world

15:09:32.103 kernel  info    stage 2048 up and running with 1 modules            application: hello-world

15:09:32.103 kernel  info    stopped stage 2048                                  application: hello-world

15:09:32.103 kernel  info    stopping stage 0                                    application: hello-world

15:09:32.103 kernel  info    kernel up and running after 3.000736083s            application: hello-world

15:09:32.103 httpserver-health-check info    trying to gracefully shutdown httpserver health check  application: hello-world

15:09:32.104 httpserver-health-check info    leaving httpserver health check                     application: hello-world

15:09:32.104 kernel  info    stopped background module http-health-check         application: hello-world

15:09:32.104 kernel  info    stopped stage 0                                     application: hello-world

15:09:32.104 kernel  info    leaving kernel with exit code 0                     application: hello-world
```

Note that the kernel waits the `wait_interval` and checks the health of the application's modules several times:

```
12:34:26.605 kernel  info    waiting for module hello-world in stage 2048 to get healthy  application: hello-world, group: health-check

12:34:27.605 kernel  info    waiting for module hello-world in stage 2048 to get healthy  application: hello-world, group: health-check

12:34:28.606 kernel  info    waiting for module hello-world in stage 2048 to get healthy  application: hello-world, group: health-check
```
