package cobrax

import "runtime/debug"

// Version is set in build step
// go build -ldflags '-X github.com/haijima/cobrax.Version=YOUR_VERSION'
var Version = ""

func version() string {
	if Version != "" {
		return Version
	}
	if buildInfo, ok := debug.ReadBuildInfo(); ok && buildInfo.Main.Version != "" {
		return buildInfo.Main.Version
	}
	return "(devel)"
}
