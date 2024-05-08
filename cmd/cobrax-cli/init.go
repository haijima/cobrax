package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/fatih/color"
	"github.com/haijima/cobrax"
	"github.com/haijima/cobrax/internal"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewInitCommand(v *viper.Viper, fs afero.Fs) *cobra.Command {
	initCmd := &cobra.Command{}
	initCmd.Use = "init [path]"
	initCmd.Aliases = []string{"new", "create", "generate"}
	initCmd.Short = "Initialize a new project"
	initCmd.Long = `Initialize a new project with the default configuration.
This command will create a new directory with the following structure:

	  project/
	  ├── .github/
	  │   └── workflows/
	  │       ├── cd.yml
	  │       └── ci.yml
	  ├── cmd/
	  │   ├── root.go
	  │   └── subcommand.go
	  ├── internal/
	  ├── .goreleaser.yaml
	  ├── .gitignore
	  ├── go.mod
	  ├── LICENSE
	  ├── main.go
	  └── README.md
`
	initCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Colorization settings
		color.NoColor = color.NoColor || v.GetBool("no-color")
		// Set Logger
		l := slog.New(slog.NewTextHandler(cmd.ErrOrStderr(), &slog.HandlerOptions{Level: cobrax.VerbosityLevel(v)}))
		slog.SetDefault(l)
		cobrax.SetLogger(l)

		// Read config file
		opts := []cobrax.ConfigOption{cobrax.WithConfigFileFlag(cmd, "config"), cobrax.WithOverrideBy(cmd.Name())}
		if err := cobrax.BindConfigs(v, cmd.Root().Name(), opts...); err != nil {
			return err
		}
		// Bind flags (flags of the command to be executed)
		if err := v.BindPFlags(cmd.Flags()); err != nil {
			return err
		}

		slog.Debug("bind flags and config values")
		slog.Debug(cobrax.DebugViper(v))
		return nil
	}
	initCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runInit(cmd, v, fs, args)
	}

	initCmd.Flags().String("name", "", "Set the name of the CLI application")
	initCmd.Flags().String("description", "", "Set the description of the project")
	initCmd.Flags().StringSlice("subcommands", nil, "Add subcommands to the root command")

	return initCmd
}

func runInit(cmd *cobra.Command, v *viper.Viper, fs afero.Fs, args []string) error {
	name := v.GetString("name")
	description := v.GetString("description")
	subcommands := v.GetStringSlice("subcommands")

	wd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "failed to get current working directory")
	}
	if len(args) > 0 && args[0] != "." {
		wd = path.Join(wd, args[0])
	}

	modInfo, err := internal.NewModInfo()
	if err != nil {
		return errors.Wrap(err, "failed to get module name")
	}

	// Prompt for missing options
	if name == "" {
		name = internal.Prompt(cmd.ErrOrStderr(), "Name", path.Base(modInfo.ModName(wd)))
	}
	if description == "" {
		description = internal.Prompt(cmd.ErrOrStderr(), "Description", "")
		if description != "" {
			description = strings.ToUpper(description[:1]) + description[1:]
		}
	}
	if len(subcommands) == 0 {
		if internal.PromptBool(cmd.ErrOrStderr(), "Use subcommands?", false) {
			subCommandNames := internal.Prompt(cmd.ErrOrStderr(), "Subcommands (comma separated)", "")
			for _, cmdName := range strings.Split(subCommandNames, ",") {
				if cmdName == "" {
					continue
				}
				subcommands = append(subcommands, strings.TrimSpace(cmdName))
			}
		}
	}
	var year int
	var author, email string
	useMIT := internal.PromptBool(cmd.ErrOrStderr(), "Use MIT license?", false)
	if useMIT {
		year = internal.PromptInt(cmd.ErrOrStderr(), "  Year for copyright", time.Now().Year())
		userName, _ := gitConfig("user.name")
		author = internal.Prompt(cmd.ErrOrStderr(), "  Author for copyright", userName)
		userEmail, _ := gitConfig("user.email")
		email = internal.Prompt(cmd.ErrOrStderr(), "  Email for copyright", userEmail)
	}
	var brewTapOwner, brewTapRepo string
	var brewDesc string
	useHomebrew := internal.PromptBool(cmd.ErrOrStderr(), "Use homebrew?", true)
	if useHomebrew {
		userName, _ := gitConfig("user.name")
		brewTapOwner = internal.Prompt(cmd.ErrOrStderr(), "  Homebrew tap repository owner", userName)
		brewTapRepo = internal.Prompt(cmd.ErrOrStderr(), "  Homebrew tap repository name", "homebrew-tap")
		brewDesc = internal.PromptWithValidate(cmd.ErrOrStderr(), "  Homebrew description", "", func(s string) error {
			s = strings.TrimSpace(s)
			if s == "" {
				return errors.New("Description should not be an empty string")
			} else if len(s) > 80 {
				return errors.New("Description is too long. It should be less than 80 characters")
			} else if strings.HasSuffix(s, ".") && !strings.HasSuffix(s, "etc.") {
				return errors.New("Description shouldn't end with a full stop")
			} else if regexp.MustCompile(`(?i)^(a|an|the) `).MatchString(s) {
				return errors.New("Description shouldn't start with an article")
			} else if regexp.MustCompile(fmt.Sprintf(`(?i)^%s\b`, name)).MatchString(s) {
				return errors.New("Description shouldn't start with the formula name")
			} else if regexp.MustCompile(`\p{So}`).MatchString(s) {
				return errors.New("Description shouldn't contain Unicode emojis or symbols")
			}
			return nil
		})
		brewDesc = strings.TrimSpace(brewDesc)
		brewDesc = strings.ToUpper(brewDesc[:1]) + brewDesc[1:]
		brewDesc = regexp.MustCompile(`c((?i)ommand ?line)`).ReplaceAllString(brewDesc, "command-line")
		brewDesc = regexp.MustCompile(`C((?i)ommand ?line)`).ReplaceAllString(brewDesc, "Command-line")
	}

	// Create project
	project := &internal.Project{
		AbsolutePath: wd,
		PkgName:      modInfo.ModName(wd),
		AppName:      name,
		Description:  description,
		SubCommands:  subcommands,
		UseMIT:       useMIT,
		CopyRight:    internal.CopyRight{Year: year, Author: author, Email: email},
		UseHomebrew:  useHomebrew,
		Homebrew:     internal.Homebrew{Description: brewDesc, TapRepository: struct{ Owner, Repo string }{Owner: brewTapOwner, Repo: brewTapRepo}},
	}
	if err := project.Create(fs); err != nil {
		return errors.Wrap(err, "failed to create project")
	}
	return nil
}

func gitConfig(key string) (string, error) {
	res, err := exec.Command("git", "config", key).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(res)), nil
}
