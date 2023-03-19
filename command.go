// Package cobrax is a thin wrapper around library [github.com/spf13/cobra] that streamlines the integration of libraries [github.com/spf13/viper] and [github.com/spf13/afero].
//
// [github.com/spf13/cobra]: http://github.com/spf13/cobra
// [github.com/spf13/viper]: http://github.com/spf13/viper
// [github.com/spf13/afero]: http://github.com/spf13/afero
package cobrax

import (
	"context"
	"fmt"
	"io"
	"log"
	"sort"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Command is a wrapper around cobra.Command that adds some additional functionality.
type Command struct {
	// Command is the underlying cobra.Command.
	*cobra.Command
	viper             *viper.Viper
	fs                afero.Fs
	parent            *Command
	commands          []*Command
	commandsAreSorted bool

	// D is the debug logger. It is set to a noop logger by default.
	D *log.Logger
	// V is the verbose logger. It is set to a noop logger by default.
	V *log.Logger

	UseConfigFile      bool
	UseEnv             bool
	UseDebugLogging    bool
	AutomaticBindViper bool

	// PersistentPreRun: children of this command will inherit and execute.
	PersistentPreRun func(cmd *Command, args []string)
	// PersistentPreRunE: PersistentPreRun but returns an error.
	PersistentPreRunE func(cmd *Command, args []string) error
	// PreRun: children of this command will not inherit.
	PreRun func(cmd *Command, args []string)
	// PreRunE: PreRun but returns an error.
	PreRunE func(cmd *Command, args []string) error
	// Run: Typically the actual work function. Most commands will only implement this.
	Run func(cmd *Command, args []string)
	// RunE: Run but returns an error.
	RunE func(cmd *Command, args []string) error
	// PostRun: run after the Run command.
	PostRun func(cmd *Command, args []string)
	// PostRunE: PostRun but returns an error.
	PostRunE func(cmd *Command, args []string) error
	// PersistentPostRun: children of this command will inherit and execute after PostRun.
	PersistentPostRun func(cmd *Command, args []string)
	// PersistentPostRunE: PersistentPostRun but returns an error.
	PersistentPostRunE func(cmd *Command, args []string) error
}

var noopLogger = log.New(io.Discard, "", 0)

// NewCommand creates a new Command.
func NewCommand(v *viper.Viper, fs afero.Fs) *Command {
	return Wrap(&cobra.Command{}, v, fs)
}

// Wrap creates a new Command from a cobra.Command.
func Wrap(cmd *cobra.Command, v *viper.Viper, fs afero.Fs) *Command {
	cmd.CompletionOptions.HiddenDefaultCmd = true
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	cmd.SilenceUsage = true // don't show help content when error occurred
	c := &Command{Command: cmd, viper: v, fs: fs}
	v.SetFs(fs)
	c.D = noopLogger
	c.V = noopLogger
	c.UseDebugLogging = true
	c.UseConfigFile = true
	c.UseEnv = true
	c.AutomaticBindViper = true
	return c
}

// Fs returns the afero.Fs filesystem.
func (c *Command) Fs() afero.Fs { return c.fs }

// Viper returns the viper.Viper instance.
func (c *Command) Viper() *viper.Viper { return c.viper }

// ExecuteContext is a wrapper around cobra.Command.ExecuteContext.
func (c *Command) ExecuteContext(ctx context.Context) error {
	c.onExecute()
	return c.Command.ExecuteContext(ctx)
}

// Execute is a wrapper around cobra.Command.Execute.
func (c *Command) Execute() error {
	c.onExecute()
	return c.Command.Execute()
}

// ExecuteContextC is a wrapper around cobra.Command.ExecuteContextC.
func (c *Command) ExecuteContextC(ctx context.Context) (*cobra.Command, error) {
	c.onExecute()
	return c.Command.ExecuteContextC(ctx)
}

// ExecuteC is a wrapper around cobra.Command.ExecuteC.
func (c *Command) ExecuteC() (cmd *cobra.Command, err error) {
	c.onExecute()
	return c.Command.ExecuteC()
}

func (c *Command) onExecute() {
	executedCommand = c
	c.addDefaultFlags()
	c.WalkCommands(delegateRunFuncs)
}

// delegateRunFuncs delegates the Run, PreRun, PostRun, and PersistentPreRun functions to those of cobra.Command.
func delegateRunFuncs(c *Command) {
	if c.RunE != nil {
		c.Command.RunE = func(cmd *cobra.Command, args []string) error {
			return c.RunE(c, args)
		}
	} else if c.Run != nil {
		c.Command.Run = func(cmd *cobra.Command, args []string) {
			c.Run(c, args)
		}
	}

	if c.PreRunE != nil {
		c.Command.PreRunE = func(cmd *cobra.Command, args []string) error {
			return c.PreRunE(c, args)
		}
	} else if c.PreRun != nil {
		c.Command.PreRun = func(cmd *cobra.Command, args []string) {
			c.PreRun(c, args)
		}
	}

	if c.PostRunE != nil {
		c.Command.PostRunE = func(cmd *cobra.Command, args []string) error {
			return c.PostRunE(c, args)
		}
	} else if c.PostRun != nil {
		c.Command.PostRun = func(cmd *cobra.Command, args []string) {
			c.PostRun(c, args)
		}
	}

	if c.PersistentPreRunE != nil {
		c.Command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			return c.PersistentPreRunE(c, args)
		}
	} else if c.PersistentPreRun != nil {
		c.Command.PersistentPreRun = func(cmd *cobra.Command, args []string) {
			c.PersistentPreRun(c, args)
		}
	}

	if c.PersistentPostRunE != nil {
		c.Command.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
			return c.PersistentPostRunE(c, args)
		}
	} else if c.PersistentPostRun != nil {
		c.Command.PersistentPostRun = func(cmd *cobra.Command, args []string) {
			c.PersistentPostRun(c, args)
		}
	}
}

func (c *Command) addDefaultFlags() {
	rootCmd := c.Root()
	if c.UseDebugLogging {
		rootCmd.PersistentFlags().Bool("debug", false, "debug level output")
		rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose level output")
		rootCmd.MarkFlagsMutuallyExclusive("debug", "verbose")
		_ = rootCmd.BindPersistentFlag("debug")
		_ = rootCmd.BindPersistentFlag("verbose")
	}
	if c.UseConfigFile {
		rootCmd.PersistentFlags().String("config", "", fmt.Sprintf("config file (default is $XDG_CONFIG_HOME/.%s.yaml)", rootCmd.Name()))
	}
}

// Commands returns a sorted slice of child commands.
func (c *Command) Commands() []*Command {
	// do not sort commands if it already sorted or sorting was disabled
	if cobra.EnableCommandSorting && !c.commandsAreSorted {
		sort.Slice(c.commands, func(i, j int) bool {
			return c.commands[i].Name() < c.commands[j].Name()
		})
		c.commandsAreSorted = true
	}
	return c.commands
}

// AddCommand adds commands to the command.
func (c *Command) AddCommand(commands ...*Command) {
	for _, cmd := range commands {
		c.Command.AddCommand(cmd.Command)
		c.commands = append(c.commands, cmd)
		c.commandsAreSorted = false
		cmd.parent = c
	}
}

// RemoveCommand removes commands from the command.
func (c *Command) RemoveCommand(cmds ...*Command) {
	for _, cmd := range cmds {
		c.removeCommand(cmd)
	}
}

func (c *Command) removeCommand(cmd *Command) {
	for i, command := range c.commands {
		if command == cmd {
			c.commands = append(c.commands[:i], c.commands[i+1:]...)
			c.commandsAreSorted = false
			cmd.parent = nil
			c.Command.RemoveCommand(cmd.Command)
			break
		}
	}
}

// ResetCommands removes all commands from the command.
func (c *Command) ResetCommands() {
	c.parent = nil
	c.commands = nil
	c.commandsAreSorted = false
	c.Command.ResetCommands()
}

// Root returns the root command.
func (c *Command) Root() *Command {
	if c.parent == nil {
		return c
	}
	return c.parent.Root()
}

// WalkCommands walks and run the given function on all commands recursively.
func (c *Command) WalkCommands(fn func(*Command)) {
	root := c.Root()
	fn(root)
	walkCommands(root, fn)
}

func walkCommands(cmd *Command, fn func(*Command)) {
	for _, c := range cmd.Commands() {
		fn(c)
		walkCommands(c, fn)
	}
}

func (c *Command) Called() bool {
	return c.CalledAs() != ""
}

// ReadFileOrStdIn returns io.ReadCloser.
// If a file is specified, it is opened and returned. Otherwise, stdin is returned.
// When a file is returned, it must be closed by the caller.
//
// Deprecated: Use OpenOrStdIn instead.
func (c *Command) ReadFileOrStdIn(fileFlag string) (io.ReadCloser, error) {
	file := c.viper.GetString(fileFlag)
	if file != "" {
		f, err := c.Fs().Open(file)
		if err != nil {
			return nil, err
		}
		return f, nil
	} else {
		return io.NopCloser(c.InOrStdin()), nil
	}
}

// OpenOrStdIn returns io.ReadCloser.
// If a filename is specified, it is opened and returned. Otherwise, stdin is returned.
// When a file is returned, it must be closed by the caller.
func (c *Command) OpenOrStdIn(filename string) (io.ReadCloser, error) {
	if filename != "" {
		f, err := c.Fs().Open(filename)
		if err != nil {
			return nil, err
		}
		return f, nil
	} else {
		return io.NopCloser(c.InOrStdin()), nil
	}
}

// BindFlag binds a flag to a viper key.
func (c *Command) BindFlag(key string) error {
	return c.viper.BindPFlag(key, c.Flags().Lookup(key))
}

// BindLocalFlag binds a local flag to a viper key.
func (c *Command) BindLocalFlag(key string) error {
	return c.viper.BindPFlag(key, c.LocalFlags().Lookup(key))
}

// BindPersistentFlag binds a persistent flag to a viper key.
func (c *Command) BindPersistentFlag(key string) error {
	return c.viper.BindPFlag(key, c.PersistentFlags().Lookup(key))
}

// BindLocalNonPersistentFlag binds a flag specific only to this to a viper key.
func (c *Command) BindLocalNonPersistentFlag(key string) error {
	return c.viper.BindPFlag(key, c.LocalNonPersistentFlags().Lookup(key))
}

// BindInheritedFlag binds a flag inherited from a parent command to a viper key.
func (c *Command) BindInheritedFlag(key string) error {
	return c.viper.BindPFlag(key, c.InheritedFlags().Lookup(key))
}

// BindNonInheritedFlag binds a flag which were not inherited from parent commands to a viper key.
func (c *Command) BindNonInheritedFlag(key string) error {
	return c.viper.BindPFlag(key, c.NonInheritedFlags().Lookup(key))
}

// BindFlags binds all flags to viper.
func (c *Command) BindFlags() error {
	return c.viper.BindPFlags(c.Flags())
}

// BindLocalFlags binds all local flags to viper.
func (c *Command) BindLocalFlags() error {
	return c.viper.BindPFlags(c.LocalFlags())
}

// BindPersistentFlags binds all persistent flags to viper.
func (c *Command) BindPersistentFlags() error {
	return c.viper.BindPFlags(c.PersistentFlags())
}

// BindLocalNonPersistentFlags binds all flags specific to only this command to viper.
func (c *Command) BindLocalNonPersistentFlags() error {
	return c.viper.BindPFlags(c.LocalNonPersistentFlags())
}

// BindInheritedFlags binds all flags inherited from a parent command to viper.
func (c *Command) BindInheritedFlags() error {
	return c.viper.BindPFlags(c.InheritedFlags())
}

// BindNonInheritedFlags binds all flags which were not inherited from parent commands to viper.
func (c *Command) BindNonInheritedFlags() error {
	return c.viper.BindPFlags(c.NonInheritedFlags())
}

// BindEnv binds a viper key to an environment variable.
func (c *Command) BindEnv(input ...string) error {
	return c.viper.BindEnv(input...)
}

// PrintOut is a convenience method to Print to the defined output, fallback to Stdout if not set.
func (c *Command) PrintOut(i ...interface{}) {
	fmt.Fprint(c.OutOrStdout(), i...)
}

// PrintOutln is a convenience method to Println to the defined output, fallback to Stdout if not set.
func (c *Command) PrintOutln(i ...interface{}) {
	c.Print(fmt.Sprintln(i...))
}

// PrintOutf is a convenience method to Printf to the defined output, fallback to Stdout if not set.
func (c *Command) PrintOutf(format string, i ...interface{}) {
	c.Print(fmt.Sprintf(format, i...))
}
