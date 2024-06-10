package cobrax

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ConfigOptions struct {
	configFile      string
	subConfigKey    string
	configFilePaths []string // File paths without extension
	configFileExts  []string
	mergeConfig     bool
}

func BindConfigs(v *viper.Viper, rootCmdName string, opts ...ConfigOption) error {
	opt := &ConfigOptions{
		configFilePaths: make([]string, 0, 12),
		configFileExts:  []string{"json", "toml", "yaml", "yml"},
		mergeConfig:     true,
	}
	rootCmdName = strings.ToLower(rootCmdName)
	xdgConfigHome := "$HOME/.config"
	if xdg, exists := os.LookupEnv("XDG_CONFIG_HOME"); exists {
		xdgConfigHome = xdg
	}
	opt.configFilePaths = append(opt.configFilePaths,
		fmt.Sprintf("%s/%s/config", xdgConfigHome, rootCmdName),
		fmt.Sprintf("$HOME/.%s", rootCmdName),
		fmt.Sprintf("./.%s", rootCmdName))

	// apply options
	for _, fn := range opts {
		fn(opt)
	}

	if opt.configFile != "" {
		v.SetConfigFile(opt.configFile) // Use config file from the flag.
		if err := v.ReadInConfig(); err != nil {
			return err
		}
		logger.Info(fmt.Sprintf("using config file: %s", v.ConfigFileUsed()))
		logger.Debug(DebugViper(v))
		// Override sub-config
		if opt.subConfigKey != "" {
			if subConf := v.GetStringMap(opt.subConfigKey); len(subConf) > 0 {
				if err := v.MergeConfigMap(subConf); err != nil {
					return err
				}
				v.Set(opt.subConfigKey, nil)
				logger.Info(fmt.Sprintf("override sub-config: %s", opt.subConfigKey))
				logger.Debug(DebugViper(v))
			}
		}
		return nil
	}
	return tryReadInConfig(v, opt)
}

func tryReadInConfig(v *viper.Viper, opt *ConfigOptions) error {
	logger.Debug("attempting to read in config file")
	found := false
	for _, cf := range opt.configFilePaths {
		for _, ext := range opt.configFileExts {
			cf, err := filepath.Abs(os.ExpandEnv(fmt.Sprintf("%s.%s", cf, ext)))
			if err != nil {
				logger.Debug(err.Error())
				continue
			}
			v.SetConfigFile(cf)
			if err = v.MergeInConfig(); err != nil {
				logger.Debug(err.Error())
				continue
			}
			logger.Info(fmt.Sprintf("successfully loaded config file: %s", v.ConfigFileUsed()))
			logger.Debug(DebugViper(v))
			found = true

			// Override sub-config
			if opt.subConfigKey != "" {
				if subConf := v.GetStringMap(opt.subConfigKey); len(subConf) > 0 {
					if err := v.MergeConfigMap(subConf); err != nil {
						return err
					}
					v.Set(opt.subConfigKey, nil)
					logger.Debug(fmt.Sprintf("override sub-config: %s", opt.subConfigKey))
					logger.Debug(DebugViper(v))
				}
			}

			if !opt.mergeConfig {
				return nil
			}
		}
	}
	if !found {
		logger.Debug("no config file found")
	}
	return nil
}

func DebugViper(v *viper.Viper) string {
	keys := v.AllKeys()
	slices.SortFunc(keys, sortConfigKey)
	buf := make([]byte, 0, 1024)
	buf = append(buf, "Config values:\n"...)
	for _, k := range keys {
		buf = append(buf, fmt.Sprintf("\t%s: %v\n", k, v.Get(k))...)
	}
	return string(buf)
}

func sortConfigKey(a, b string) int {
	if a == b {
		return 0
	}
	aK, aV, aHasDot := strings.Cut(a, ".")
	bK, bV, bHasDot := strings.Cut(b, ".")
	if aHasDot == bHasDot {
		if aK == bK {
			return sortConfigKey(aV, bV)
		}
		return strings.Compare(aK, bK)
	}
	if aHasDot {
		return 1
	} else {
		return -1
	}
}

// <editor-fold desc="ConfigOptions">

type ConfigOption func(*ConfigOptions)

func WithConfigFileName(file string) ConfigOption {
	return func(opt *ConfigOptions) {
		opt.configFile = file
	}
}

func WithConfigFileFlag(cmd *cobra.Command, flagName string) ConfigOption {
	return func(opt *ConfigOptions) {
		opt.configFile = cmd.Flag(flagName).Value.String()
	}
}

func WithConfigFilePaths(paths ...string) ConfigOption {
	return func(opt *ConfigOptions) {
		opt.configFilePaths = paths
	}
}

func WithOverrideBy(key string) ConfigOption {
	return func(opt *ConfigOptions) {
		opt.subConfigKey = strings.ToLower(key)
	}
}

func WithOverrideDisabled() ConfigOption {
	return func(opt *ConfigOptions) {
		opt.subConfigKey = ""
	}
}

func WithMergeConfig(merge bool) ConfigOption {
	return func(opt *ConfigOptions) {
		opt.mergeConfig = merge
	}
}

//</editor-fold>
