package internal

import _ "embed"

//go:embed templates/main.go.tmpl
var mainGoTemplate string

//go:embed templates/cmd/root.go.tmpl
var rootGoTemplate string

//go:embed templates/cmd/subcmd.go.tmpl
var subcmdGoTemplate string

//go:embed templates/.gitignore.tmpl
var gitIgnoreTemplate string

//go:embed templates/LICENSE.tmpl
var mitLicenseTemplate string

//go:embed templates/.github/workflows/ci.yaml.tmpl
var ciYamlTemplate string

//go:embed templates/.github/workflows/release.yaml.tmpl
var releaseYamlTemplate string

//go:embed templates/.goreleaser.yaml.tmpl
var goreleaserYamlTemplate string
