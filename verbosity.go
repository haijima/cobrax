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
	return slog.LevelError - slog.Level(4*VerbosityCount(v))
}
