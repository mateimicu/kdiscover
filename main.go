package main

import (
	"flag"

	"github.com/mateimicu/kdiscover/cmd"
)

//nolint:unused, varcheck, deadcode
var update = flag.Bool("update", false, "update .golden files")

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.Execute(version, commit, date)
}
