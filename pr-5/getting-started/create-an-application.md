# Create an application

For running a simple gosoline application we use the following 3 files:

* `module.go` → kernel module with business logic
* `config.dist.yml` → base configuration file
* `main.go` → setup and run the application

## Create the kernel module[​](#create-the-kernel-module "Direct link to Create the kernel module")

The kernel module contains the actual application logic. The `HelloWorldModule` struct has to implement the `kernel.Module` interface. The func `NewHelloWorldModule` is of the type `kernel.ModuleFactory` and is used to add the module to the application.

module.go

```
package main



import (

	"context"



	"github.com/justtrackio/gosoline/pkg/cfg"

	"github.com/justtrackio/gosoline/pkg/kernel"

	"github.com/justtrackio/gosoline/pkg/log"

)



func NewHelloWorldModule(ctx context.Context, config cfg.Config, logger log.Logger) (kernel.Module, error) {

	return &HelloWorldModule{

		logger: logger.WithChannel("hello-world"),

	}, nil

}



type HelloWorldModule struct {

	logger log.Logger

}



func (h HelloWorldModule) Run(ctx context.Context) error {

	h.logger.Info(ctx, "Hello World")



	return nil

}
```

## Create the main file[​](#create-the-main-file "Direct link to Create the main file")

In the main function, we're running the application by adding the `HelloWorldModule` to the kernel via the `NewHelloWorldModule` factory.

main.go

```
package main



import "github.com/justtrackio/gosoline/pkg/application"



func main() {

	application.Run(

		application.WithConfigFile("config.dist.yml", "yml"),

		application.WithModuleFactory("hello-world", NewHelloWorldModule),

	)

}
```

## Create the config file[​](#create-the-config-file "Direct link to Create the config file")

The last missing piece for a minimal gosoline application is the config file. You load it explicitly using `application.WithConfigFile("config.dist.yml", "yml")`. For a minimal version it has at least the application identifiers and enviroment value.

config.dist.yml

```
app:

  env: dev

  name: hello-world

  namespace: "{app.tags.project}.{app.env}.{app.tags.family}.{app.tags.group}"

  tags:

    project: gosoline

    family: get-started

    group: grp
```

## Application output[​](#application-output "Direct link to Application output")

Running the application will result in the following output:

stdout

```
15:09:17.803 main    info    applied priority 8 config post processor 'gosoline.log.handler_main'  application: hello-world

15:09:17.803 main    info    applied priority 1 config post processor 'gosoline.dx.useRandomPort'  application: hello-world

15:09:17.803 main    info    applied priority 1 config post processor 'gosoline.reslife.autoCreate'  application: hello-world

15:09:17.803 kernel  info    starting kernel                                     application: hello-world

15:09:17.803 kernel  info    cfg app.env=dev                                     application: hello-world

15:09:17.803 kernel  info    cfg app.name=hello-world                            application: hello-world

15:09:17.803 kernel  info    cfg app.namespace={app.tags.project}.{app.env}.{app.tags.family}.{app.tags.group}  application: hello-world

15:09:17.803 kernel  info    cfg app.tags.family=get-started                     application: hello-world

15:09:17.803 kernel  info    cfg app.tags.group=grp                              application: hello-world

15:09:17.803 kernel  info    cfg app.tags.project=gosoline                       application: hello-world

15:09:17.803 kernel  info    cfg appctx.metadata.server.port=0                   application: hello-world

15:09:17.803 kernel  info    cfg dx.use_random_port=true                         application: hello-world

15:09:17.803 kernel  info    cfg httpserver.health-check.port=0                  application: hello-world

15:09:17.803 kernel  info    cfg kernel.health_check.timeout=1m0s                application: hello-world

15:09:17.803 kernel  info    cfg kernel.health_check.wait_interval=10ms          application: hello-world

15:09:17.803 kernel  info    cfg kernel.kill_timeout=30s                         application: hello-world

15:09:17.803 kernel  info    cfg kinsumer.autoscale.dynamodb.naming.pattern={app.env}-kinsumer-autoscale-leaders  application: hello-world

15:09:17.803 kernel  info    cfg kinsumer.autoscale.ecs.client=default           application: hello-world

15:09:17.803 kernel  info    cfg kinsumer.autoscale.ecs.cluster={app.env}        application: hello-world

15:09:17.803 kernel  info    cfg kinsumer.autoscale.ecs.service={app.name}       application: hello-world

15:09:17.803 kernel  info    cfg kinsumer.autoscale.enabled=false                application: hello-world

15:09:17.803 kernel  info    cfg kinsumer.autoscale.leader_election=kinsumer-autoscale  application: hello-world

15:09:17.803 kernel  info    cfg kinsumer.autoscale.orchestrator=ecs             application: hello-world

15:09:17.803 kernel  info    cfg kinsumer.autoscale.period=1m0s                  application: hello-world

15:09:17.803 kernel  info    cfg log.handlers.main.channels.kernel.level=        application: hello-world

15:09:17.803 kernel  info    cfg log.handlers.main.channels.main.level=          application: hello-world

15:09:17.803 kernel  info    cfg log.handlers.main.formatter=console             application: hello-world

15:09:17.803 kernel  info    cfg log.handlers.main.level=info                    application: hello-world

15:09:17.803 kernel  info    cfg log.handlers.main.timestamp_format=15:04:05.000  application: hello-world

15:09:17.803 kernel  info    cfg log.handlers.main.type=iowriter                 application: hello-world

15:09:17.803 kernel  info    cfg log.handlers.main.writer=stdout                 application: hello-world

15:09:17.803 kernel  info    cfg resource_lifecycles.create.enabled=true         application: hello-world

15:09:17.803 kernel  info    cfg fingerprint: ec1a0765339d8ad26b6f13ae9b10514c   application: hello-world

15:09:17.803 kernel  info    stage 2048 up and running with 1 modules            application: hello-world

15:09:17.803 kernel  info    kernel up and running after 415µs                   application: hello-world

15:09:17.803 kernel  info    running foreground module hello-world in stage 2048  application: hello-world

15:09:17.804 hello-world info    Hello World                                         application: hello-world

15:09:17.804 kernel  info    stopped foreground module hello-world               application: hello-world

15:09:17.804 kernel  info    stopping kernel due to: no more foreground modules in running state  application: hello-world

15:09:17.804 kernel  info    stopping stage 2048                                 application: hello-world

15:09:17.804 kernel  info    stopped stage 2048                                  application: hello-world

15:09:17.804 kernel  info    leaving kernel with exit code 0                     application: hello-world
```
