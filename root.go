package cobrax

import (
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewRoot returns the base command used when called without any subcommands
func NewRoot(v *viper.Viper) *cobra.Command {
	rootCmd := &cobra.Command{}
	rootCmd.Args = cobra.NoArgs
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.SilenceUsage = true  // don't show help content when error occurred
	rootCmd.SilenceErrors = true // Print error by own slog logger
	rootCmd.Version = VersionFunc()

	rootCmd.PersistentFlags().BoolP("version", "V", false, "Print version information and quit")
	rootCmd.PersistentFlags().String("config", "", "config file (default is $XDG_CONFIG_HOME/.stool.yaml)")
	rootCmd.PersistentFlags().Bool("no-color", false, "disable colorized output")
	_ = v.BindPFlag("no-color", rootCmd.PersistentFlags().Lookup("no-color"))
	rootCmd.PersistentFlags().CountP("verbose", "v", "More output per occurrence. (Use -vvvv or --verbose 4 for max verbosity)")
	_ = v.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Silence all output")
	_ = v.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	rootCmd.MarkFlagsMutuallyExclusive("verbose", "quiet")

	return rootCmd
}

func RootPersistentPreRunE(cmd *cobra.Command, v *viper.Viper, args []string) error {
	// Colorize settings (Do before logger setup)
	color.NoColor = color.NoColor || v.GetBool("no-color")

	// Read config file
	if err := NewConfigBinder(cmd).Bind(v); err != nil {
		return err
	}
	if err := OverrideBySubConfig(v, strings.ToLower(cmd.Name())); err != nil {
		return err
	}
	// Bind flags (flags of the command to be executed)
	if err := v.BindPFlags(cmd.Flags()); err != nil {
		return err
	}
	// Print config values
	logger.Info("bind flags and config values")
	logger.Debug(DebugViper(v))

	return nil
}
