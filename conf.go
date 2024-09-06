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
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Deprecated != "" || f.Hidden || f.Name == "help" {
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
		b, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return err
		}
		_, err = w.Write(b)
		return err
	}
	return nil
}
