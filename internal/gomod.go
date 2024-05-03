package internal

import (
	"encoding/json"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/cockroachdb/errors"
)

type ModInfo struct {
	Path, Dir string
}

func NewModInfo() (*ModInfo, error) {
	var modInfo ModInfo
	m, err := exec.Command("go", "list", "-json", "-m").Output()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current directory")
	}
	if err := json.Unmarshal(m, &modInfo); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal current directory")
	}
	if modInfo.Path == "command-line-arguments" {
		return nil, errors.New("Please run `go mod init <MODNAME>` before `cobrax-cli init`")
	}
	return &modInfo, nil
}

func (m *ModInfo) ModName(wd string) string {
	rel := strings.Split(strings.TrimPrefix(wd, m.Dir), string(filepath.Separator))
	return path.Join(slices.Insert(rel, 0, m.Path)...)
}

func GoGet(mod string) error {
	return exec.Command("go", "get", mod).Run()
}

func GoModTidy() error {
	return exec.Command("go", "mod", "tidy").Run()
}
