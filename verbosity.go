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

func VerbosityLevel(v *viper.Viper, opts ...VerbosityLevelOption) slog.Level {
	options := &VerbosityLevelOptions{zeroLevel: slog.LevelError, step: 4}
	for _, o := range opts {
		o(options)
	}
	return options.zeroLevel - slog.Level(options.step*VerbosityCount(v))
}

type VerbosityLevelOptions struct {
	zeroLevel slog.Level
	step      int
}

type VerbosityLevelOption func(*VerbosityLevelOptions)

func WithVerbosityZeroLevel(level slog.Level) VerbosityLevelOption {
	return func(o *VerbosityLevelOptions) {
		o.zeroLevel = level
	}
}

func WithVerbosityStep(step int) VerbosityLevelOption {
	return func(o *VerbosityLevelOptions) {
		o.step = step
	}
}
