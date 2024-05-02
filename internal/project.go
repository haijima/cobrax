package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/cockroachdb/errors"
	"github.com/spf13/afero"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Project contains name, license and paths to projects.
type Project struct {
	PkgName      string
	AbsolutePath string
	AppName      string
	Description  string
	SubCommands  []string
}

func (p *Project) Create(fs afero.Fs) error {
	// Create main.go
	if err := createFile(fs, fmt.Sprintf("%s/main.go", p.AbsolutePath), MainTemplate(), p); err != nil {
		return errors.Wrap(err, "failed to create main.go")
	}

	// Create cmd/root.go
	if err := createFile(fs, fmt.Sprintf("%s/cmd/root.go", p.AbsolutePath), RootTemplate(), p); err != nil {
		return errors.Wrap(err, "failed to create cmd/root.go")
	}

	// Create cmd/subcommand.go
	for _, subCommand := range p.SubCommands {
		if err := createFile(fs, fmt.Sprintf("%s/cmd/%s.go", p.AbsolutePath, subCommand), SubCommandTemplate(), map[string]string{"CmdName": subCommand}); err != nil {
			return errors.Wrapf(err, "failed to create cmd/%s.go", subCommand)
		}
	}

	// Create internal/
	if err := createFile(fs, fmt.Sprintf("%s/internal/.gitkeep", p.AbsolutePath), "", nil); err != nil {
		return errors.Wrap(err, "failed to create internal/.gitkeep")
	}

	// Create README.md
	if err := createFile(fs, fmt.Sprintf("%s/README.md", p.AbsolutePath), "", nil); err != nil {
		return errors.Wrap(err, "failed to create README.md")
	}

	// Create .gitignore
	if err := createFile(fs, fmt.Sprintf("%s/.gitignore", p.AbsolutePath), GitIgnoreTemplate(), nil); err != nil {
		return errors.Wrap(err, "failed to create .gitignore")
	}

	// go get and go mod tidy
	mods := []string{
		"github.com/fatih/color",
		"github.com/haijima/cobrax",
		"github.com/spf13/afero",
		"github.com/spf13/cobra",
		"github.com/spf13/viper",
	}
	for _, mod := range mods {
		if err := GoGet(mod); err != nil {
			return errors.Wrapf(err, "failed to go get %s", mod)
		}
	}
	return GoModTidy()
}

var funcMap = template.FuncMap{
	"title": cases.Title(language.Und).String,
}

func createFile(fs afero.Fs, path string, tpl string, data any) error {
	// create parent directory
	if err := createDirectoryIfNotExists(fs, filepath.Dir(path)); err != nil {
		return err
	}
	// create file
	file, err := fs.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	// write template
	return template.Must(template.New(path).Funcs(funcMap).Parse(tpl)).Execute(file, data)
}

func createDirectoryIfNotExists(fs afero.Fs, path string) error {
	if _, err := fs.Stat(path); os.IsNotExist(err) {
		return fs.MkdirAll(path, 0754)
	}
	return nil
}
