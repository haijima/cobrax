package internal

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/mattn/go-isatty"
)

type prompter struct {
	Message  string
	Default  string
	Validate func(string) error
}

func (p *prompter) prompt(w io.Writer) string {
	if !(isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())) ||
		!(isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())) {
		return p.Default
	}

	fmt.Fprint(w, p.Message+": ")
	var input string
	scanner := bufio.NewScanner(os.Stdin)
	if ok := scanner.Scan(); ok {
		input = strings.TrimRight(scanner.Text(), "\r\n")
	}
	if input == "" {
		input = p.Default
	}
	if err := p.Validate(input); err != nil {
		fmt.Fprintln(w, err.Error())
		return p.prompt(w) // Try again
	}
	return input
}

// Prompt shows simple prompt
func Prompt(w io.Writer, message, defaultAnswer string) string {
	msg := message
	if defaultAnswer != "" {
		msg += fmt.Sprintf(" [%s]", defaultAnswer)
	}
	return (&prompter{
		Message:  msg,
		Default:  defaultAnswer,
		Validate: func(input string) error { return nil },
	}).prompt(w)
}

// Confirm shows a yes/no prompt
func Confirm(w io.Writer, message string, defaultToYes bool) bool {
	confirm := "y/N"
	if defaultToYes {
		confirm = "Y/n"
	}
	input := (&prompter{
		Message: fmt.Sprintf("%s [%s]", message, confirm),
		Validate: func(input string) error {
			if input != "" && input != "y" && input != "Y" && input != "n" && input != "N" {
				return errors.New("Enter 'y' or 'n'")
			}
			return nil
		},
	}).prompt(w)
	return input == "y" || input == "Y" || (input == "" && defaultToYes)
}
