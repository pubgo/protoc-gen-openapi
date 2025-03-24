package version

import (
	"fmt"
	"runtime"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func FullVersion() string {
	return fmt.Sprintf("%s (%s) @ %s; %s", version, commit, date, runtime.Version())
}
