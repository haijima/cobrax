package cobrax

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

var format PrintConfigFormat

func PrintConfigCmd(name string) *cobra.Command {
	genConfCmd := &cobra.Command{}
	genConfCmd.Use = name
	genConfCmd.Short = "Generate configuration file"
	genConfCmd.Args = cobra.NoArgs
	genConfCmd.RunE = func(cmd *cobra.Command, _ []string) error {
		return PrintConfig(cmd.OutOrStdout(), GetFlags(cmd.Root()), format)
	}

	genConfCmd.Flags().Var(&format, "format", "The output format {toml|yaml|json}")

	return genConfCmd
}

func GetFlags(cmd *cobra.Command) map[string]any {
	m := make(map[string]any)
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if f.Deprecated != "" || f.Hidden || f.Name == "help" || f.Name == "version" {
			return
		}
		m[f.Name] = f.Value.String()
	})

	for _, c := range cmd.Commands() {
		if c.Deprecated != "" || c.Hidden {
			continue
		}
		child := GetFlags(c)
		if len(child) > 0 {
			m[c.Name()] = child
		}
	}
	return m
}

func PrintConfig(w io.Writer, m map[string]any, format PrintConfigFormat) error {
	switch format {
	case YAML:
		return yaml.NewEncoder(w).Encode(m)
	case TOML:
		return toml.NewEncoder(w).Encode(m)
	case JSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(m)
	}
	return nil
}

type PrintConfigFormat string

const (
	YAML PrintConfigFormat = "yaml"
	JSON PrintConfigFormat = "json"
	TOML PrintConfigFormat = "toml"
)

// String is used both by fmt.Print and by Cobra in help text
func (f *PrintConfigFormat) String() string {
	return string(*f)
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (f *PrintConfigFormat) Set(v string) error {
	switch v {
	case "yaml", "json", "toml":
		*f = PrintConfigFormat(v)
		return nil
	default:
		return errors.New(`must be one of "yaml", "json", or "toml"`)
	}
}

// Type is only used in help text
func (f *PrintConfigFormat) Type() string {
	return "format"
}
