package cobrax

import (
	"strings"

	"github.com/spf13/pflag"
)

// SnakeToKebab normalizes flag names from snake_case to kebab-case.
func SnakeToKebab(f *pflag.FlagSet, name string) pflag.NormalizedName {
	return pflag.NormalizedName(strings.Replace(name, "_", "-", -1))
}
