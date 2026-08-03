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

If you want to have more control over the module's health status, you can implement the `kernel.HealthChecked` interface for your application modules.

## Customize your health check[​](#customize-your-health-check "Direct link to Customize your health check")

To control how the health of your application is verified, you'll implement the `kernel.HealthChecked` interface:

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
13:34:26.600 main    info    applied priority 8 config post processor 'gosoline.log.handler_main'  application: hello-world, group: health-check

13:34:26.601 main    info    applied priority 1 config post processor 'gosoline.dx.autoCreate'  application: hello-world, group: health-check

13:34:26.601 main    info    applied priority 1 config post processor 'gosoline.dx.useRandomPort'  application: hello-world, group: health-check

12:34:26.604 kernel  info    starting kernel                                     application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg api.health.path=/health                         application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg api.health.port=8090                            application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg app_family=how-to                               application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg app_group=health-check                          application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg app_name=hello-world                            application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg app_project=gosoline                            application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg appctx.metadata.server.port=0                   application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg dx.auto_create=true                             application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg dx.use_random_port=true                         application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg env=dev                                         application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg kernel.health_check.timeout=10s                 application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg kernel.health_check.wait_interval=1s            application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg kernel.kill_timeout=10s                         application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg log.handlers.main.formatter=console             application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg log.handlers.main.level=info                    application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg log.handlers.main.timestamp_format=15:04:05.000  application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg log.handlers.main.type=iowriter                 application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg log.handlers.main.writer=stdout                 application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg metric.application={app_name}                   application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg metric.cloudwatch.naming.pattern={project}/{env}/{family}/{group}-{app}  application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg metric.enabled=false                            application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg metric.environment={env}                        application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg metric.family={app_family}                      application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg metric.group={app_group}                        application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg metric.interval=1m0s                            application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg metric.project={app_project}                    application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg metric.writer=                                  application: hello-world, group: health-check

12:34:26.604 kernel  info    cfg fingerprint: ef5386d483b4b86e621859e38ecf2442   application: hello-world, group: health-check

12:34:26.604 kernel  info    stage 0 up and running with 1 modules               application: hello-world, group: health-check

12:34:26.604 kernel  info    running background module metric in stage 0         application: hello-world, group: health-check

12:34:26.605 kernel  info    running background module metadata-server in stage 1024  application: hello-world, group: health-check

12:34:26.605 kernel  info    running background module api-health-check in stage 1024  application: hello-world, group: health-check

12:34:26.605 metrics info    metrics not enabled..                               application: hello-world, group: health-check

12:34:26.605 kernel  info    stopped background module metric                    application: hello-world, group: health-check

12:34:26.605 kernel  info    stage 1024 up and running with 2 modules            application: hello-world, group: health-check

12:34:26.605 kernel  info    waiting for module hello-world in stage 2048 to get healthy  application: hello-world, group: health-check

12:34:26.605 kernel  info    running foreground module hello-world in stage 2048  application: hello-world, group: health-check

12:34:26.605 metadata-server info    serving metadata on address [::]:44891              application: hello-world, group: health-check

12:34:27.605 kernel  info    waiting for module hello-world in stage 2048 to get healthy  application: hello-world, group: health-check

12:34:28.606 kernel  info    waiting for module hello-world in stage 2048 to get healthy  application: hello-world, group: health-check

12:34:29.606 hello-world info    Hello World                                         application: hello-world, group: health-check

12:34:29.606 kernel  info    stopped foreground module hello-world               application: hello-world, group: health-check

12:34:29.606 kernel  info    stage 2048 up and running with 1 modules            application: hello-world, group: health-check

12:34:29.606 kernel  info    kernel up and running                               application: hello-world, group: health-check

12:34:29.606 kernel  info    stopping kernel due to: no more foreground modules in running state  application: hello-world, group: health-check

12:34:29.606 kernel  info    stopping stage 2048                                 application: hello-world, group: health-check

12:34:29.606 kernel  info    stopped stage 2048                                  application: hello-world, group: health-check

12:34:29.606 kernel  info    stopping stage 1024                                 application: hello-world, group: health-check

12:34:29.607 kernel  info    stopped background module api-health-check          application: hello-world, group: health-check

12:34:29.607 kernel  info    stopped background module metadata-server           application: hello-world, group: health-check

12:34:29.607 kernel  info    stopped stage 1024                                  application: hello-world, group: health-check

12:34:29.607 kernel  info    stopping stage 0                                    application: hello-world, group: health-check

12:34:29.607 kernel  info    stopped stage 0                                     application: hello-world, group: health-check

12:34:29.607 kernel  info    leaving kernel with exit code 0                     application: hello-world, group: health-check
```

Note that the kernel waits the `wait_interval` and checks the health of the application's modules several times:

```
12:34:26.605 kernel  info    waiting for module hello-world in stage 2048 to get healthy  application: hello-world, group: health-check

12:34:27.605 kernel  info    waiting for module hello-world in stage 2048 to get healthy  application: hello-world, group: health-check

12:34:28.606 kernel  info    waiting for module hello-world in stage 2048 to get healthy  application: hello-world, group: health-check
```
