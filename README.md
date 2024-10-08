# cobrax

[![CI Status](https://github.com/haijima/cobrax/workflows/CI/badge.svg?branch=main)](https://github.com/haijima/cobrax/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/haijima/cobrax.svg)](https://pkg.go.dev/github.com/haijima/cobrax)
[![Go report](https://goreportcard.com/badge/github.com/haijima/cobrax)](https://goreportcard.com/report/github.com/haijima/cobrax)

A utility library for [spf13/cobra](http://github.com/spf13/cobra), [spf13/viper](http://github.com/spf13/viper)
and [spf13/afero](http://github.com/spf13/afero).

## Usage

```go
var version, commit, date string // main.version, main.commit, main.date

cmd := cobrax.NewRoot(viper.New())
cmd.Use = "app"
cmd.Short = "description of app"
cmd.Version = cobrax.VersionFunc(version, commit, date)

cmd.AddCommand(someCmd)
cmd.AddCommand(otherCmd)

cmd.Execute()
```

```go
// Open the file. When pipe is used and the filename is empty, read from stdin.
cobrax.OpenOrStdIn(viper.GetString("filename"), afero.NewOsFs()) 
```

```go
// Read the config file and set the values to the viper.
cobrax.BindConfigs(v, "app")
```

## License

This tool is licensed under the MIT License. See the [LICENSE](https://github.com/haijima/cobrax/blob/main/LICENSE) file
for details.

## Links

- [spf13/cobra](http://github.com/spf13/cobra)
- [spf13/viper](http://github.com/spf13/viper)
- [spf13/afero](http://github.com/spf13/afero)
