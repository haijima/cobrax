package main

import (
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/fatih/color"
	"github.com/haijima/cobrax"
	"github.com/haijima/cobrax/internal"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewInitCommand(v *viper.Viper, fs afero.Fs) *cobra.Command {
	initCmd := &cobra.Command{}
	initCmd.Use = "init [path]"
	initCmd.Short = "Initialize a new project"
	initCmd.Long = `Initialize a new project with the default configuration.
This command will create a new directory with the following structure:

	  project/
	  ├── cmd/
	  │   ├── root.go
	  │   └── subcommand.go
	  ├── internal/
	  ├── .gitignore
	  ├── go.mod
	  ├── main.go
	  └── README.md
`
	initCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
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
	initCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runInit(cmd, v, fs, args)
	}

	initCmd.Flags().String("name", "", "Set the name of the CLI application")
	initCmd.Flags().String("description", "", "Set the description of the project")
	initCmd.Flags().StringSlice("subcommands", nil, "Add subcommands to the root command")

	return initCmd
}

func runInit(cmd *cobra.Command, v *viper.Viper, fs afero.Fs, args []string) error {
	name := v.GetString("name")
	description := v.GetString("description")
	subcommands := v.GetStringSlice("subcommands")

	wd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "failed to get current working directory")
	}
	if len(args) > 0 && args[0] != "." {
		wd = path.Join(wd, args[0])
	}

	modInfo, err := internal.NewModInfo()
	if err != nil {
		return errors.Wrap(err, "failed to get module name")
	}

	// Prompt for missing options
	if name == "" {
		name = internal.Prompt(cmd.ErrOrStderr(), "Name", path.Base(modInfo.ModName(wd)))
	}
	if description == "" {
		description = internal.Prompt(cmd.ErrOrStderr(), "Description", "")
		if description != "" {
			description = strings.ToUpper(description[:1]) + description[1:]
		}
	}
	if len(subcommands) == 0 {
		if internal.Confirm(cmd.ErrOrStderr(), "Use subcommands?", false) {
			subCommandNames := internal.Prompt(cmd.ErrOrStderr(), "Subcommands (comma separated)", "")
			for _, cmdName := range strings.Split(subCommandNames, ",") {
				if cmdName == "" {
					continue
				}
				subcommands = append(subcommands, strings.TrimSpace(cmdName))
			}
		}
	}

	// Create project
	project := &internal.Project{
		AbsolutePath: wd,
		PkgName:      modInfo.ModName(wd),
		AppName:      name,
		Description:  description,
		SubCommands:  subcommands,
	}
	if err := project.Create(fs); err != nil {
		return errors.Wrap(err, "failed to create project")
	}
	return nil
}
