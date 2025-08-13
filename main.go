package main

import (
	"github.com/sdcio/config-diff/cmd"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	cmd.Execute()
}
