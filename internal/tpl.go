package internal

func MainTemplate() string {
	return `package main

import (
	"fmt"
	"log/slog"
	"os"

	"{{ .PkgName }}/cmd"
	"github.com/haijima/cobrax"
	"github.com/mattn/go-colorable"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

// https://goreleaser.com/cookbooks/using-main.version/
var version, commit, date string

func main() {
	v := viper.NewWithOptions(viper.WithLogger(slog.Default()))
	fs := afero.NewOsFs()
	v.SetFs(fs)
	rootCmd := cmd.NewRootCmd(v, fs)
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
`
}

func RootTemplate() string {
	return `package cmd

import (
	"log/slog"

	"github.com/fatih/color"
	"github.com/haijima/cobrax"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewRootCmd represents the root command of CLI
func NewRootCmd(v *viper.Viper, fs afero.Fs) *cobra.Command {
	rootCmd := cobrax.NewRoot(v)
	rootCmd.Use = "{{ .AppName }}"
	rootCmd.Short = "{{ .Description }}"
{{- if not .SubCommands }}
	rootCmd.Args = nil
{{- end }}
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Colorization settings
		color.NoColor = color.NoColor || v.GetBool("no-color")
		// Set Logger
		l := slog.New(slog.NewTextHandler(cmd.ErrOrStderr(), &slog.HandlerOptions{Level: cobrax.VerbosityLevel(v)}))
		slog.SetDefault(l)
		cobrax.SetLogger(l)

		// Read config file
		opts := []cobrax.ConfigOption{cobrax.WithConfigFileFlag(cmd, "config"), cobrax.WithOverrideBy(cmd.Name())}
		if err := cobrax.BindConfigs(v, cmd.Root().Name(), opts...); err != nil {
			return err
		}
		// Bind flags (flags of the command to be executed)
		if err := v.BindPFlags(cmd.Flags()); err != nil {
			return err
		}

		slog.Debug("bind flags and config values")
		slog.Debug(cobrax.DebugViper(v))
		return nil
	}
{{- if not .SubCommands }}
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runRoot(cmd, v, fs, args)
	}
{{- else }}
{{ range .SubCommands }}
	rootCmd.AddCommand(New{{ . | title }}Cmd(v, fs)){{ end }}
{{- end }}

	rootCmd.SetGlobalNormalizationFunc(cobrax.SnakeToKebab)

	return rootCmd
}
{{- if not .SubCommands }}

func runRoot(cmd *cobra.Command, v *viper.Viper, fs afero.Fs, args []string) error {
	return nil
}
{{- end }}
`
}

func SubCommandTemplate() string {
	return `package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// New{{ .CmdName | title }}Cmd represents the {{ .CmdName }} command
func New{{ .CmdName | title }}Cmd(v *viper.Viper, fs afero.Fs) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Use = "{{ .CmdName }}"
	cmd.Short = "Description for {{ .CmdName }} command"
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return run{{ .CmdName | title }}(cmd, v, fs, args)
	}

	// You can add flags here
	//cmd.Flags().StringP("name", "n", "", "name")

	return cmd
}

func run{{ .CmdName | title }}(cmd *cobra.Command, v *viper.Viper, fs afero.Fs, args []string) error {
	return nil
}
`
}

func GitIgnoreTemplate() string {
	return `
go.work
*.test
*.out
out/
.idea/*
.DS_Store
`
}
