package cobrax

import "github.com/spf13/cobra"

var executedCommand *Command

func init() {
	cobra.OnInitialize(func() {
		if executedCommand == nil {
			return // Executed via cobra.Command.Execute() not cobrax.Command.Execute()
		}
		root := executedCommand.Root()

		var ancestryForCalledCommand func(Command, []Command) []Command
		ancestryForCalledCommand = func(parent Command, path []Command) []Command {
			newPath := append([]Command{}, path...)
			newPath = append(newPath, parent)

			if parent.CalledAs() != "" {
				return newPath
			}

			for _, child := range parent.Commands() {
				foundPath := ancestryForCalledCommand(*child, newPath)
				if foundPath != nil {
					return foundPath
				}
			}

			return nil
		}

		for _, ancestor := range ancestryForCalledCommand(*root, []Command{}) {
			if ancestor.AutomaticBindViper {
				cobra.CheckErr(ancestor.BindFlags())
			}
		}
	})
}
