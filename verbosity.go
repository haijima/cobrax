package cobrax

import (
	"log/slog"

	"github.com/spf13/viper"
)

func verbosityCount(v *viper.Viper, verboseName, quietName string) int {
	if v.GetBool(quietName) {
		return -1
	}
	return v.GetInt(verboseName)
}

func VerbosityLevel(v *viper.Viper, opts ...VerbosityLevelOption) slog.Level {
	options := &VerbosityLevelOptions{verboseName: "verbose", quietName: "quiet", zeroLevel: slog.LevelError, step: 4}
	for _, o := range opts {
		o(options)
	}
	return options.zeroLevel - slog.Level(options.step*verbosityCount(v, options.verboseName, options.quietName))
}

type VerbosityLevelOptions struct {
	verboseName string
	quietName   string
	zeroLevel   slog.Level
	step        int
}

type VerbosityLevelOption func(*VerbosityLevelOptions)

func WithVerbosityName(verboseName string) VerbosityLevelOption {
	return func(o *VerbosityLevelOptions) {
		o.verboseName = verboseName
	}
}

func WithQuietName(quietName string) VerbosityLevelOption {
	return func(o *VerbosityLevelOptions) {
		o.quietName = quietName
	}
}

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
