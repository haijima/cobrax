package cobrax

import (
	"github.com/spf13/cobra"
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
		for _, ancestor := range ancestryForCalledCommand(*root, nil) {
			for _, initializer := range initializers {
				initializer(&ancestor)
			}
		}
	})

	// bind flags to viper on cobrax.OnInitialize
	OnInitialize(func(cmd *Command) {
		if cmd.AutomaticBindViper {
			cobra.CheckErr(cmd.BindFlags())
		}
	})
}

// OnInitialize sets the passed functions to be run when each command's
// Execute method is called.
func OnInitialize(y ...Initializer) {
	initializers = append(initializers, y...)
}
