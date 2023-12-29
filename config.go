package cobrax

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ConfigBinder struct {
	cmd                *cobra.Command
	configFlagName     string
	configFile         string
	globalConfigFiles  []string // File paths without extension
	projectConfigFiles []string // File paths without extension
}

func NewConfigBinder(cmd *cobra.Command) *ConfigBinder {
	cb := &ConfigBinder{cmd: cmd}
	cb.configFlagName = "config"
	cb.configFile = ""
	rootCmdName := strings.ToLower(cmd.Root().Name())

	cb.globalConfigFiles = make([]string, 0, 8)
	if xdg, exists := os.LookupEnv("XDG_CONFIG_HOME"); exists {
		cb.globalConfigFiles = append(cb.globalConfigFiles, fmt.Sprintf("%s/%s/config.json", xdg, rootCmdName))
		cb.globalConfigFiles = append(cb.globalConfigFiles, fmt.Sprintf("%s/%s/config.toml", xdg, rootCmdName))
		cb.globalConfigFiles = append(cb.globalConfigFiles, fmt.Sprintf("%s/%s/config.yaml", xdg, rootCmdName))
		cb.globalConfigFiles = append(cb.globalConfigFiles, fmt.Sprintf("%s/%s/config.yml", xdg, rootCmdName))
	} else {
		cb.globalConfigFiles = append(cb.globalConfigFiles, fmt.Sprintf("$HOME/.config/%s/config.json", rootCmdName))
		cb.globalConfigFiles = append(cb.globalConfigFiles, fmt.Sprintf("$HOME/.config/%s/config.toml", rootCmdName))
		cb.globalConfigFiles = append(cb.globalConfigFiles, fmt.Sprintf("$HOME/.config/%s/config.yaml", rootCmdName))
		cb.globalConfigFiles = append(cb.globalConfigFiles, fmt.Sprintf("$HOME/.config/%s/config.yml", rootCmdName))
	}
	cb.globalConfigFiles = append(cb.globalConfigFiles, fmt.Sprintf("$HOME/.%s.json", rootCmdName))
	cb.globalConfigFiles = append(cb.globalConfigFiles, fmt.Sprintf("$HOME/.%s.toml", rootCmdName))
	cb.globalConfigFiles = append(cb.globalConfigFiles, fmt.Sprintf("$HOME/.%s.yaml", rootCmdName))
	cb.globalConfigFiles = append(cb.globalConfigFiles, fmt.Sprintf("$HOME/.%s.yml", rootCmdName))
	cb.projectConfigFiles = []string{
		fmt.Sprintf("./.%s.json", rootCmdName),
		fmt.Sprintf("./.%s.toml", rootCmdName),
		fmt.Sprintf("./.%s.yaml", rootCmdName),
		fmt.Sprintf("./.%s.yml", rootCmdName),
	}
	return cb
}

func NewConfigBinderWithOption(cmd *cobra.Command, option ...ConfigOption) *ConfigBinder {
	cb := NewConfigBinder(cmd)
	for _, opt := range option {
		cb = opt.apply(cb)
	}
	return cb
}

func (b *ConfigBinder) Bind(v *viper.Viper, fs afero.Fs) error {
	configFile := b.configFile
	if configFile == "" {
		configFile = b.cmd.Flags().Lookup(b.configFlagName).Value.String()
	}
	if configFile != "" {
		v.SetConfigFile(configFile) // Use config file from the flag.
		if err := v.ReadInConfig(); err != nil {
			return err
		}
		logger.Info(fmt.Sprintf("using config file: %s", v.ConfigFileUsed()))
		logger.Debug(DebugViper(v))
	} else {
		_, globalOk := tryReadInConfig(v, fs, b.globalConfigFiles)
		_, projectOk := tryReadInConfig(v, fs, b.projectConfigFiles)
		if globalOk && projectOk {
			logger.Info("merge global and project config files")
			logger.Debug(DebugViper(v))
		}
	}
	return nil
}

func tryReadInConfig(v *viper.Viper, fs afero.Fs, files []string) (*viper.Viper, bool) {
	logger.Info("attempting to read in config file")
	for _, cf := range files {
		cf, err := filepath.Abs(os.ExpandEnv(cf))
		if err != nil {
			logger.Debug(err.Error())
			continue
		}
		vv := viper.New()
		vv.SetFs(fs)
		vv.SetConfigFile(cf)
		err = vv.ReadInConfig()
		logger.Debug("reading file", "file", cf)
		if err == nil {
			logger.Info(fmt.Sprintf("successfully loaded config file: %s", vv.ConfigFileUsed()))
			logger.Debug(DebugViper(vv))
			if err := v.MergeConfigMap(vv.AllSettings()); err != nil {
				logger.Debug(err.Error())
				return v, false
			}
			v.SetConfigFile(vv.ConfigFileUsed())
			return v, true
		}
	}
	logger.Debug("no config file found")
	return v, false
}

func OverrideBySubConfig(v *viper.Viper, key string) error {
	if subConf := v.GetStringMap(key); subConf != nil {
		if err := v.MergeConfigMap(subConf); err != nil {
			return err
		}
		logger.Info(fmt.Sprintf("override sub-config: %s", key))
		logger.Debug(DebugViper(v))
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
	ah, at, aok := strings.Cut(a, ".")
	bh, bt, bok := strings.Cut(b, ".")
	if aok == bok {
		if ah == bh {
			return sortConfigKey(at, bt)
		} else if ah < bh {
			return -1
		} else {
			return 1
		}
	}
	if aok {
		return 1
	} else {
		return -1
	}
}

// <editor-fold desc="ConfigOptions">

type ConfigOption interface {
	apply(*ConfigBinder) *ConfigBinder
}

type configOptionFunc func(*ConfigBinder) *ConfigBinder

func (fn configOptionFunc) apply(b *ConfigBinder) *ConfigBinder {
	return fn(b)
}

func WithConfigFile(file string) ConfigOption {
	return configOptionFunc(func(b *ConfigBinder) *ConfigBinder {
		b.configFile = file
		return b
	})
}

func WithConfigFlagName(name string) ConfigOption {
	return configOptionFunc(func(b *ConfigBinder) *ConfigBinder {
		b.configFlagName = name
		return b
	})
}

func WithGlobalConfigFiles(paths ...string) ConfigOption {
	return configOptionFunc(func(b *ConfigBinder) *ConfigBinder {
		b.globalConfigFiles = paths
		return b
	})
}

func WithProjectConfigFiles(paths ...string) ConfigOption {
	return configOptionFunc(func(b *ConfigBinder) *ConfigBinder {
		b.projectConfigFiles = paths
		return b
	})
}

//</editor-fold>
