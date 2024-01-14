package main

import (
	"github.com/sev-2/raiden/cmd/cli/generate"
	"github.com/sev-2/raiden/cmd/cli/start"

	"github.com/sev-2/raiden/cmd/cli/run"
	"github.com/sev-2/raiden/cmd/cli/version"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{Use: "raiden"}
	rootCmd.AddCommand(
		generate.Command(),
		start.Command(),
		run.Command(),
		version.Command(),
	)
	rootCmd.Execute()
}
