# cfg package

With the cfg package, you can configure and use configurations in your gosoline project.

## Functions[​](#functions "Direct link to Functions")

### [WithConfigFile()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/cfg/options.go#L23)[​](#withconfigfile "Direct link to withconfigfile")

#### Usage[​](#usage "Direct link to Usage")

```
options := []cfg.Option{

    cfg.WithConfigFile("config.dist.yml", "yml"),

}
```

#### Description[​](#description "Direct link to Description")

Load configurations from a file.

info

YAML is currently the only supported configuration filetype.

### GetSettings() and GetDefaults()[​](#getsettings-and-getdefaults "Direct link to GetSettings() and GetDefaults()")

Available in Gosoline **v0.68.3** and later, these generic helpers return a settings value together with an error:

```
type Settings struct {

    Port int `cfg:"port" default:"8080" validate:"min=1,max=65535"`

}



settings, err := cfg.GetSettings[Settings](config, "server")

if err != nil {

    return err

}



defaults, err := cfg.GetDefaults[Settings](config)

if err != nil {

    return err

}
```

`GetSettings` delegates to `config.UnmarshalKey`, including configured values, environment overrides, defaults, and validation. `GetDefaults` delegates to `config.UnmarshalDefaults`; it reads defaults without loading configured values. Both accept additional `cfg.UnmarshalDefaults` arguments. See the [released implementations](https://github.com/justtrackio/gosoline/blob/v0.68.3/pkg/cfg/config.go).

## Methods[​](#methods "Direct link to Methods")

### [GetInt()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/cfg/config.go#L129)[​](#getint "Direct link to getint")

#### Usage[​](#usage-1 "Direct link to Usage")

With default:

```
config := cfg.New()

config.GetInt("num", 1)
```

Without default:

```
config := cfg.New()



options := []cfg.Option{

    cfg.WithConfigSetting("num", 1),

}



if err := config.Option(options...); err != nil {

    panic(err)

}



config.GetInt("num")
```

#### Description[​](#description-1 "Direct link to Description")

Gets an integer from the configuration struct.

caution

Environment variables overwrite configuration values:

```
config := cfg.New()



options := []cfg.Option{

    cfg.WithConfigSetting("port", "8080"),

}



if err := config.Option(options...); err != nil {

    panic(err)

}



port, err := config.GetString("port")

if err != nil {

    panic(err)

}

fmt.Println(port) // Prints 8080



os.Setenv("PORT", "8081")



port, err = config.GetString("port")

if err != nil {

    panic(err)

}

fmt.Println(port) // Prints 8081
```

#### Related methods[​](#related-methods "Direct link to Related methods")

GetString()

```
config := cfg.New()

str, err := config.GetString("env", "dev")

if err != nil {

    panic(err)

}
```

GetBool()

```
config := cfg.New()

enabled, err := config.GetBool("enabled", true)
```

## Variables[​](#variables "Direct link to Variables")

### [DefaultEnvKeyReplacer](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/cfg/config.go#L62)[​](#defaultenvkeyreplacer "Direct link to defaultenvkeyreplacer")

#### Usage[​](#usage-2 "Direct link to Usage")

```
options := []cfg.Option{

    cfg.WithEnvKeyReplacer(cfg.DefaultEnvKeyReplacer),

}
```

#### Description[​](#description-2 "Direct link to Description")

Replaces "." and "-" with "\_" for env key loading.
