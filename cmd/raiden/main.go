package main

import (
	"github.com/sev-2/raiden/cmd/raiden/commands"
	"github.com/sev-2/raiden/pkg/cli"

	"github.com/spf13/cobra"
)

func main() {
	f := cli.LogFlags{}

	rootCmd := &cobra.Command{Use: "raiden"}

	rootCmd.AddCommand(
		commands.ApplyCommand(),
		commands.BuildCommand(),
		commands.ConfigureCommand(),
		commands.GenerateCommand(),
		commands.ImportCommand(),
		commands.InitCommand(),
		commands.RunCommand(),
		commands.ServeCommand(),
		commands.StartCommand(),
		commands.VersionCommand(),
	)

	f.Bind(rootCmd)

	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
