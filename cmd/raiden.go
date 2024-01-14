package main

import (
	"github.com/sev-2/raiden/cmd/generate"
	"github.com/sev-2/raiden/cmd/new"

	"github.com/sev-2/raiden/cmd/run"
	"github.com/sev-2/raiden/cmd/version"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{Use: "raiden"}
	rootCmd.AddCommand(
		generate.Command(),
		new.Command(),
		run.Command(),
		version.Command(),
	)
	rootCmd.Execute()
}
