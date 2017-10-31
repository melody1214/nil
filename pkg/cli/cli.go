package cli

import (
	"github.com/chanyoung/nil/pkg/cmd"
)

// Main is the entry point for the cli.
func Main() {
	cmd.RootCmd.Execute()
}
