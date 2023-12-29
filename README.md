# cobrax

[![CI Status](https://github.com/haijima/cobrax/workflows/CI/badge.svg?branch=main)](https://github.com/haijima/cobrax/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/haijima/cobrax.svg)](https://pkg.go.dev/github.com/haijima/cobrax)
[![Go report](https://goreportcard.com/badge/github.com/haijima/cobrax)](https://goreportcard.com/report/github.com/haijima/cobrax)

A utility library for [spf13/cobra](http://github.com/spf13/cobra), [spf13/viper](http://github.com/spf13/viper) and [spf13/afero](http://github.com/spf13/afero).

## Usage

```go
// Build as follows to set the version.
// go build -ldflags '-X github.com/haijima/cobrax.Version=YOUR_VERSION'
cmd := &cobra.Command{
    Use: "app",
	Version: cobrax.VersionFunc(),
}
```

```go
filename := viper.GetString("filename")
cobrax.OpenOrStdIn(filename, afero.NewOsFs(), cmd.InOrStdin()) // Open the file if exsits, otherwise return os.Stdin.
```

```go
// Read the config file and set the values to the viper.
cobrax.NewConfigBinder(cmd).Bind(v)
```

```go
// Override config value by sub-command specific config.
cobrax.OverrideBySubConfig(v, strings.ToLower(cmd.Name())
```

## License

This tool is licensed under the MIT License. See the [LICENSE](https://github.com/haijima/cobrax/blob/main/LICENSE) file
for details.

## Links

- [spf13/cobra](http://github.com/spf13/cobra)
- [spf13/viper](http://github.com/spf13/viper)
- [spf13/afero](http://github.com/spf13/afero)
