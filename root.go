package cobrax

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewRoot returns the base command used when called without any subcommands
func NewRoot(v *viper.Viper) *cobra.Command {
	return NewRootWithOption(v, DefaultRootFlagOption)
}

// NewRootWithOption returns the base command used when called without any subcommands
func NewRootWithOption(v *viper.Viper, option RootFlagOption) *cobra.Command {
	rootCmd := &cobra.Command{}
	rootCmd.Args = cobra.NoArgs
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.SilenceUsage = true  // don't show help content when error occurred
	rootCmd.SilenceErrors = true // Print error by own slog logger

	if option.Version.Name != "" {
		rootCmd.PersistentFlags().BoolP(option.Version.Name, option.Version.Shorthand, false, option.Version.Usage)
	}
	if option.Config.Name != "" {
		rootCmd.PersistentFlags().StringP(option.Config.Name, option.Config.Shorthand, "", option.Config.Usage)
	}
	if option.NoColor.Name != "" {
		rootCmd.PersistentFlags().BoolP(option.NoColor.Name, option.NoColor.Shorthand, false, option.NoColor.Usage)
		_ = v.BindPFlag(option.NoColor.Name, rootCmd.PersistentFlags().Lookup(option.NoColor.Name))
	}
	if option.Verbose.Name != "" {
		rootCmd.PersistentFlags().CountP(option.Verbose.Name, option.Verbose.Shorthand, option.Verbose.Usage)
		_ = v.BindPFlag(option.Verbose.Name, rootCmd.PersistentFlags().Lookup(option.Verbose.Name))
	}
	if option.Quiet.Name != "" {
		rootCmd.PersistentFlags().BoolP(option.Quiet.Name, option.Quiet.Shorthand, false, option.Quiet.Usage)
		_ = v.BindPFlag(option.Quiet.Name, rootCmd.PersistentFlags().Lookup(option.Quiet.Name))
	}
	if option.Verbose.Name != "" && option.Quiet.Name != "" {
		rootCmd.MarkFlagsMutuallyExclusive(option.Verbose.Name, option.Quiet.Name)
	}

	return rootCmd
}

type RootFlagOption struct {
	Version FlagOption
	Config  FlagOption
	NoColor FlagOption
	Verbose FlagOption
	Quiet   FlagOption
}

type FlagOption struct {
	Name      string
	Shorthand string
	Usage     string
}

var DefaultRootFlagOption = RootFlagOption{
	Version: FlagOption{Name: "version", Shorthand: "V", Usage: "Print version information and quit"},
	Config:  FlagOption{Name: "config", Shorthand: "", Usage: "configuration `filename`"},
	NoColor: FlagOption{Name: "no-color", Shorthand: "", Usage: "disable colorized output"},
	Verbose: FlagOption{Name: "verbose", Shorthand: "v", Usage: "More output per occurrence. (e.g. -vvv)"},
	Quiet:   FlagOption{Name: "quiet", Shorthand: "q", Usage: "Silence all output"},
}

func RootPersistentPreRunE(cmd *cobra.Command, v *viper.Viper, _ afero.Fs, _ []string) error {
	// Read config file
	opts := []ConfigOption{WithConfigFileFlag(cmd, "config"), WithOverrideBy(cmd.Name())}
	if err := BindConfigs(v, cmd.Root().Name(), opts...); err != nil {
		return err
	}
	// Bind flags (flags of the command to be executed)
	if err := v.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	logger.Debug("bind flags and config values")
	logger.Debug(DebugViper(v))
	return nil
}
