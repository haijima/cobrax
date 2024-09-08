package cobrax

import (
	"encoding/json"
	"io"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

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

type PrintConfigFormat int

const (
	YAML PrintConfigFormat = iota
	TOML
	JSON
)

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

func GenConfigRunE(format PrintConfigFormat) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		flags := GetFlags(cmd.Root())
		return PrintConfig(cmd.OutOrStdout(), flags, YAML)
	}
}
