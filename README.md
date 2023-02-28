# cobrax

[![CI Status](https://github.com/haijima/cobrax/workflows/CI/badge.svg?branch=main)](https://github.com/haijima/cobrax/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/haijima/cobrax.svg)](https://pkg.go.dev/github.com/haijima/cobrax)
[![Go report](https://goreportcard.com/badge/github.com/haijima/cobrax)](https://goreportcard.com/report/github.com/haijima/cobrax)

A thin wrapper around library [spf13/cobra](http://github.com/spf13/cobra) that streamlines the integration of library [spf13/viper](http://github.com/spf13/viper) and [spf13/afero](http://github.com/spf13/afero).

## Usage

```go
package main

import (
	"fmt"
	"os"

	"github.com/haijima/cobrax"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

func main() {
	rootCmd := cobrax.NewCommand(viper.New(), afero.NewOsFs())
	rootCmd.Use = "test-cli"
	rootCmd.Short = "cobrax sample CLI"
	rootCmd.Run = func(cmd *cobrax.Command, args []string) {
		fmt.Println("test-cli called")
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

## Installation

```shell
go get github.com/haijima/cobrax@latest
```

## License

This tool is licensed under the MIT License. See the [LICENSE](https://github.com/haijima/cobrax/blob/main/LICENSE) file
for details.

## Links

- [spf13/cobra](http://github.com/spf13/cobra)
- [spf13/viper](http://github.com/spf13/viper)
- [spf13/afero](http://github.com/spf13/afero)
