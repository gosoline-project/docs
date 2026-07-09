# Load configurations

There are different ways to load configurations into your application:

From a file

First, configure your project in a yaml file:

config.dist.yml

```
port: 80

host: example.com



struct_example:

  port: 80

  host: example.com

  timeout: 5s

  threshold: 100
```

Then, load your configurations into your application:

main.go

```
package main



import (

	// 1

	"github.com/justtrackio/gosoline/pkg/cfg"

)



func main() {

	// 2

	config := cfg.New()



	// 3

	options := []cfg.Option{

		cfg.WithConfigFile("config.dist.yml", "yml"),

	}



	// 4

	if err := config.Option(options...); err != nil {

		panic(err)

	}

}
```

Here, you:

1. Import the `cfg` package.
2. Create a new configuration object.
3. Load configurations from the file.
4. Apply the options.

From a map

main.go

```
package main



import (

	// 1

	"github.com/justtrackio/gosoline/pkg/cfg"

)



func main() {

	// 2

	config := cfg.New()



	// 3

	options := []cfg.Option{

		cfg.WithConfigMap(map[string]any{

			"enabled": true,

		}),

	}



	// 4

	if err := config.Option(options...); err != nil {

		panic(err)

	}

}
```

Here, you:

1. Import the `cfg` package.
2. Create a new configuration object.
3. Load configurations from the map.
4. Apply the options.

From individual settings

main.go

```
package main



import (

	// 1

	"github.com/justtrackio/gosoline/pkg/cfg"

)



func main() {

	// 2

	config := cfg.New()



	// 3

	options := []cfg.Option{

		cfg.WithConfigSetting("host", "localhost"),

		cfg.WithConfigSetting("port", "80"),

	}



	// 4

	if err := config.Option(options...); err != nil {

		panic(err)

	}

}
```

Here, you've:

1. Import the `cfg` package.
2. Create a new configuration object.
3. Load configurations from settings.
4. Apply the options.

From multiple sources

main.go

```
package main



import (

	// 1

	"github.com/justtrackio/gosoline/pkg/cfg"

)



func main() {

	// 2

	config := cfg.New()



	// 3

	options := []cfg.Option{

		cfg.WithConfigFile("config.dist.yml", "yml"),

		cfg.WithConfigMap(map[string]any{

			"enabled": true,

		}),

		cfg.WithConfigSetting("foo", "bar"),

	}



	// 4

	if err := config.Option(options...); err != nil {

		panic(err)

	}

}
```

Here, you've:

1. Import the `cfg` package.
2. Create a new configuration object.
3. Load configurations from multiple sources.
4. Apply the options.

Once you've loaded your configurations, you can access them with type-specific getters. For example:

main.go

```
package main



import (

	"fmt"



	"github.com/justtrackio/gosoline/pkg/cfg"

)



func main() {

	config := cfg.New()



	options := []cfg.Option{

		cfg.WithConfigFile("config.dist.yml", "yml"),

	}



	if err := config.Option(options...); err != nil {

		panic(err)

	}



	port, err := config.GetInt("port")

	if err != nil {

		panic(fmt.Errorf("failed to get port: %w", err))

	}

	fmt.Printf("got port: %d\n", port)



	host, err := config.GetString("host")

	if err != nil {

		panic(fmt.Errorf("failed to get host: %w", err))

	}

	fmt.Printf("got host: %s\n", host)

}
```

## Conclusion[​](#conclusion "Direct link to Conclusion")

In this guide, you've learned multiple ways to load configurations into your app with gosoline.

Check out these resources to learn more about configurations and gosoline:

* [Configure a logger with gosoline](/docs/pr-1/how-to/logging/use-loggers.md#implement-a-logger-with-an-app-configuration)
* [API reference for the cfg package](/docs/pr-1/reference/package-cfg.md)
