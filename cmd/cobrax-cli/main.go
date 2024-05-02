package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/haijima/cobrax"
	"github.com/mattn/go-colorable"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

// https://goreleaser.com/cookbooks/using-main.version/
var version, commit, date string

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(colorable.NewColorableStderr(), &slog.HandlerOptions{Level: slog.LevelWarn})))
	v := viper.NewWithOptions(viper.WithLogger(slog.Default()))
	fs := afero.NewOsFs()
	v.SetFs(fs)
	rootCmd := NewRootCmd(v, fs)
	rootCmd.Version = cobrax.VersionFunc(version, commit, date)
	rootCmd.SetOut(colorable.NewColorableStdout())
	rootCmd.SetErr(colorable.NewColorableStderr())
	if err := rootCmd.Execute(); err != nil {
		if slog.Default().Enabled(rootCmd.Context(), slog.LevelDebug) {
			slog.Error(fmt.Sprintf("%+v", err))
		} else {
			slog.Error(err.Error())
		}
		os.Exit(1)
	}
}
