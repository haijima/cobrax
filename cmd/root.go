package main

import (
	"log/slog"

	"github.com/fatih/color"
	"github.com/haijima/cobrax"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewRootCmd returns the base command used when called without any subcommands
func NewRootCmd(v *viper.Viper, fs afero.Fs) *cobra.Command {
	rootCmd := cobrax.NewRoot(v)
	rootCmd.Use = "cobrax-cli"
	rootCmd.Short = "cobrax-cli is a simple command-line tool to create a CLI project with cobrax"
	rootCmd.SetGlobalNormalizationFunc(cobrax.SnakeToKebab)
	rootCmd.Args = cobra.RangeArgs(0, 1)
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Colorization settings
		color.NoColor = color.NoColor || v.GetBool("no-color")
		// Set Logger
		l := slog.New(slog.NewTextHandler(cmd.ErrOrStderr(), &slog.HandlerOptions{Level: cobrax.VerbosityLevel(v)}))
		slog.SetDefault(l)
		cobrax.SetLogger(l)

		return cobrax.RootPersistentPreRunE(cmd, v, fs, args)
	}

	rootCmd.AddCommand(NewInitCommand(v, fs))

	return rootCmd
}
