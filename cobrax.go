package cobrax

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Initializer func(*Command)

var initializers []Initializer
var executedCommand *Command

func init() {
	// Hook cobrax.initializers into cobra.OnInitialize
	cobra.OnInitialize(func() {
		if executedCommand == nil {
			return // Executed via cobra.Command.Execute() not cobrax.Command.Execute()
		}
		root := executedCommand.Root()

		var ancestryForCalledCommand func(Command, []Command) []Command
		ancestryForCalledCommand = func(cmd Command, path []Command) []Command {
			newPath := append(path, cmd)
			if cmd.Called() {
				return newPath
			}
			for _, child := range cmd.Commands() {
				if p := ancestryForCalledCommand(*child, newPath); p != nil {
					return p
				}
			}
			return nil
		}

		// Run initializers on all ancestors for the called command
		ancestors := ancestryForCalledCommand(*root, nil)
		for _, initializer := range initializers {
			for _, ancestor := range ancestors {
				initializer(&ancestor)
			}
		}
	})

	OnInitialize(bindFlags, bindConfigFile, prioritizeNestedConfigValue, useDebugLogger, bindEnv)
}

// OnInitialize sets the passed functions to be run when each command's
// Execute method is called.
func OnInitialize(y ...Initializer) {
	initializers = append(initializers, y...)
}

func bindFlags(cmd *Command) {
	if cmd.AutomaticBindViper {
		cobra.CheckErr(cmd.BindFlags())
	}
}

// bindConfigFile binds the config file to the command.
func bindConfigFile(cmd *Command) {
	if cmd.HasParent() || !cmd.UseConfigFile {
		return
	}

	if cmd.Flag("config").Changed {
		cmd.viper.SetConfigFile(cmd.Flag("config").Value.String()) // Use config file from the flag.
	} else {
		wd, err := os.Getwd()
		cobra.CheckErr(err)
		cmd.viper.AddConfigPath(wd) // adding current working directory as first search path
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig == "" {
			home, err := os.UserHomeDir()
			cobra.CheckErr(err)
			xdgConfig = filepath.Join(home, ".config")
		}
		cmd.viper.AddConfigPath(filepath.Join(xdgConfig, cmd.Name())) // adding XDG config directory as second search path
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		cmd.viper.AddConfigPath(home) // adding home directory as third search path
		cmd.viper.SetConfigName("." + cmd.Name())
	}

	// If a config file is found, read it in.
	if err := cmd.viper.ReadInConfig(); err == nil {
		cmd.D.Printf("Using config file: %s", cmd.viper.ConfigFileUsed())
		cmd.V.Printf("%+v", cmd.viper.AllSettings())
	}
}

// prioritizeNestedConfigValue overrides config values with those that have the subcommand prefix.
func prioritizeNestedConfigValue(cmd *Command) {
	if !cmd.Called() || !cmd.HasParent() {
		return
	}

	filename := cmd.viper.ConfigFileUsed()
	if filename == "" {
		return
	}
	file, err := afero.ReadFile(cmd.Fs(), filename)
	cobra.CheckErr(err)

	configViper := viper.New()
	configViper.SetConfigFile(filename)
	err = configViper.ReadConfig(bytes.NewReader(file))
	cobra.CheckErr(err)

	prefix := strings.Join(strings.Split(cmd.CommandPath(), " ")[1:], ".") + "."
	overrideMap := make(map[string]interface{})
	for _, k := range configViper.AllKeys() {
		if strings.HasPrefix(k, prefix) {
			targetKey := strings.Replace(k, prefix, "", 1)
			if strings.Contains(targetKey, ".") {
				continue
			}
			overrideMap[targetKey] = configViper.Get(k)
		}
	}
	err = cmd.viper.MergeConfigMap(overrideMap)
	cobra.CheckErr(err)
}

func useDebugLogger(cmd *Command) {
	if !cmd.UseDebugLogging {
		return
	}

	if cmd.viper.GetBool("debug") && cmd.D == noopLogger {
		cmd.D = log.New(cmd.ErrOrStderr(), "", 0)
	} else if cmd.viper.GetBool("verbose") {
		if cmd.D == noopLogger {
			cmd.D = log.New(cmd.ErrOrStderr(), "[DEBUG]   ", log.Ldate|log.Ltime|log.Llongfile)
		}
		if cmd.V == noopLogger {
			cmd.V = log.New(cmd.ErrOrStderr(), "[VERBOSE] ", log.Ldate|log.Ltime|log.Llongfile)
		}
	}
}

func bindEnv(cmd *Command) {
	if cmd.HasParent() || !cmd.UseEnv {
		return
	}

	cmd.viper.SetEnvPrefix(strings.ToUpper(cmd.Name()))
	cmd.viper.AutomaticEnv() // read in environment variables that match
}
