package run

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/cmd/cli/generate"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

type Flags struct {
	ConfigPath   string
	ProjectPath  string
	MainFilePath string
}

func Command() *cobra.Command {
	flags := Flags{}

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Deploy and run application",
		Long:  "Deploy resource , run backend application",
		RunE:  runCmd(&flags),
	}

	cmd.Flags().StringVarP(&flags.MainFilePath, "file", "f", "", "Path to the main file")
	cmd.Flags().StringVarP(&flags.ConfigPath, "config", "c", "", "Path to the configuration file")
	cmd.Flags().StringVarP(&flags.ProjectPath, "project", "p", "", "Path to project folder")

	return cmd
}

func runCmd(flags *Flags) func(*cobra.Command, []string) error {
	// set default value
	if flags.ConfigPath == "" {
		flags.ConfigPath = "./configs/app.yaml"
	}

	return func(cmd *cobra.Command, args []string) error {
		// load config
		config, err := raiden.LoadConfig(&flags.ConfigPath)
		if err != nil {
			return err
		}

		if flags.ProjectPath != "" {
			config.ProjectName = flags.ProjectPath
		}

		// 1. TODO : compare resource

		// 2. regenerate resource
		if err := generate.GenerateResource(config); err != nil {
			return err
		}

		// 3. build app
		mainFileName := utils.ToSnakeCase(config.ProjectName)
		mainFile := fmt.Sprintf("/cmd/%s/%s.go", config.ProjectName, mainFileName)
		outputFile, err := build(mainFile)
		if err != nil {
			return err
		}

		// 4. TODO : run deployment if in local

		// 5. run app
		return run(outputFile)
	}
}

func build(file string) (string, error) {
	filePath, err := utils.GetAbsolutePath(file)
	if err != nil {
		return "", err
	}

	// Determine the output binary name based on the Go file name
	outputBinary := "raiden"
	if runtime.GOOS == "windows" {
		outputBinary += ".exe"
	}

	outputDir := "build"
	output := filepath.Join(outputDir, outputBinary)

	// Run the "go build" command
	cmd := exec.Command("go", "build", "-o", output, filePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error building binary: %v", err)
	}

	return output, nil
}

func run(binaryPath string) error {
	absBinaryPath, err := utils.GetAbsolutePath(binaryPath)
	if err != nil {
		return err
	}

	cmd := exec.Command(absBinaryPath)

	// Redirect standard input, output, and error to the current process
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error running binary : %v", err)
	}

	return nil

}
