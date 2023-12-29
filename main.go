package cobrax

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version is set in build step
// go build -ldflags '-X github.com/haijima/cobrax.Version=YOUR_VERSION'
var Version = ""

var logger *slog.Logger

func init() {
	logger = slog.Default()
}

func SetLogger(l *slog.Logger) {
	logger = l
}

func VersionFunc() string {
	if Version != "" {
		return Version
	}
	if buildInfo, ok := debug.ReadBuildInfo(); ok && buildInfo.Main.Version != "" {
		return buildInfo.Main.Version
	}
	return "(devel)"
}

func OpenOrStdIn(filename string, fs afero.Fs, stdin io.Reader) (io.ReadCloser, error) {
	if filename != "" {
		f, err := fs.Open(filename)
		if err != nil {
			return nil, err
		}
		return f, nil
	} else {
		return io.NopCloser(stdin), nil
	}
}

func ReadConfigFile(v *viper.Viper, cfgFile string, override bool, subCommandName string) error {
	if cfgFile != "" {
		v.SetConfigFile(cfgFile) // Use config file from the flag.
		// If a config file is found, read it in.
		if err := v.ReadInConfig(); err == nil {
			logger.Info(fmt.Sprintf("Using config file: %s", v.ConfigFileUsed()))
			logger.Debug(DebugViper(v))
		}
	} else {
		xdgViper := viper.NewWithOptions(viper.WithLogger(logger))
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			xdgConfig = filepath.Join(home, ".config")
		}
		xdgViper.AddConfigPath(filepath.Join(xdgConfig, "stool")) // use XDG config directory as global config path
		xdgViper.SetConfigName("config")

		homeViper := viper.NewWithOptions(viper.WithLogger(logger))
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		homeViper.AddConfigPath(home) // use home directory as global config path
		homeViper.SetConfigName(".stool")

		projectViper := viper.NewWithOptions(viper.WithLogger(logger))
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		projectViper.AddConfigPath(wd) // use current working directory as project config path
		projectViper.SetConfigName(".stool")

		// Read in global config file (XDG > HOME)
		if err := xdgViper.ReadInConfig(); err == nil {
			logger.Info(fmt.Sprintf("Using file as global config: %s", xdgViper.ConfigFileUsed()))
			logger.Debug(DebugViper(xdgViper))
		} else if err := homeViper.ReadInConfig(); err == nil {
			logger.Info(fmt.Sprintf("Using file as global config: %s", homeViper.ConfigFileUsed()))
			logger.Debug(DebugViper(homeViper))
		}

		// Read in project config file
		if err := projectViper.ReadInConfig(); err == nil {
			logger.Info(fmt.Sprintf("Using file as project config: %s", projectViper.ConfigFileUsed()))
			logger.Debug(DebugViper(projectViper))
		}

		// Merge all config files
		if (xdgViper.ConfigFileUsed() != "" || homeViper.ConfigFileUsed() != "") && projectViper.ConfigFileUsed() != "" {
			cobra.CheckErr(v.MergeConfigMap(homeViper.AllSettings()))
			cobra.CheckErr(v.MergeConfigMap(xdgViper.AllSettings()))
			cobra.CheckErr(v.MergeConfigMap(projectViper.AllSettings()))
			logger.Info("Global config and project config are merged")
			logger.Debug(DebugViper(v))
		}
	}

	if override {
		if subConf := v.GetStringMap(strings.ToLower(subCommandName)); subConf != nil {
			if err := v.MergeConfigMap(subConf); err != nil {
				return err
			}
			logger.Info(fmt.Sprintf("Override sub-command specific config: %s", subCommandName))
			logger.Debug(DebugViper(v))
		}
	}
	return nil
}

func DebugViper(v *viper.Viper) string {
	keys := v.AllKeys()
	slices.SortFunc(keys, func(a, b string) int {
		if a == b {
			return 0
		}
		if strings.Contains(a, ".") && !strings.Contains(b, ".") {
			return 1
		} else if !strings.Contains(a, ".") && strings.Contains(b, ".") {
			return -1
		}
		if a < b {
			return -1
		} else {
			return 1
		}
	})
	buf := make([]byte, 0, 1024)
	buf = append(buf, "Config values:\n"...)
	for _, k := range keys {
		buf = append(buf, fmt.Sprintf("  %s: %v\n", k, v.Get(k))...)
	}
	return string(buf)
}
