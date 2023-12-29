package cobrax

import (
	"log/slog"

	"github.com/spf13/viper"
)

func VerbosityCount(v *viper.Viper) int {
	if v.GetBool("quiet") {
		return -1
	}
	return v.GetInt("verbose")
}

func VerbosityLevel(v *viper.Viper) slog.Level {
	i := VerbosityCount(v)
	return slog.Level(8 - 4*i)
}
