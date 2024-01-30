package run

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/cmd/cli/generate"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

type Flags struct {
	ConfigPath   string
	ProjectPath  string
	MainFilePath string
	Verbose      bool
}

func Command() *cobra.Command {
	flags := Flags{}

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Deploy and run application",
		Long:  "Deploy resource, run backend application",
		RunE:  runCmd(&flags),
	}

	cmd.Flags().StringVarP(&flags.MainFilePath, "file", "f", "", "Path to the main file")
	cmd.Flags().StringVarP(&flags.ConfigPath, "config", "c", "", "Path to the configuration file")
	cmd.Flags().StringVarP(&flags.ProjectPath, "project", "p", "", "Path to project folder")
	cmd.PersistentFlags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Enable verbose mode")

	return cmd
}

func runCmd(flags *Flags) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// set default value
		curDir, err := utils.GetCurrentDirectory()
		if err != nil {
			return err
		}

		if flags.ConfigPath == "" {
			flags.ConfigPath = filepath.Join(curDir, "configs/app.yaml")
		}

		if flags.ProjectPath == "" {
			flags.ProjectPath = curDir
		}

		if flags.Verbose {
			logger.SetDebug()
		}

		// load config
		logger.Debug("Load configuration from : ", flags.ConfigPath)
		config, err := raiden.LoadConfig(&flags.ConfigPath)
		if err != nil {
			return err
		}

		// 1. TODO : compare resource

		// 2. regenerate resource
		if err := generate.GenerateResource(flags.ProjectPath, config); err != nil {
			return err
		}

		// 3. build app
		mainFileName := utils.ToSnakeCase(config.ProjectName)
		mainFile := fmt.Sprintf("cmd/%s/%s.go", config.ProjectName, mainFileName)
		outputFile, err := build(flags.ProjectPath, mainFileName, mainFile)
		if err != nil {
			return err
		}

		// 4. TODO : run deployment if in local

		// 5. run app
		return run(outputFile)
	}
}

func build(projectPath string, outputBinary string, file string) (string, error) {
	// Determine the output binary name based on the Go file name
	if runtime.GOOS == "windows" {
		outputBinary += ".exe"
	}

	// set abs file path
	filePath := filepath.Join(projectPath, file)

	// set output
	outputDir := "build"
	output := filepath.Join(projectPath, outputDir, outputBinary)

	// Run the "go build" command
	logger.Debugf("Execute command : go build -o %s %s", output, filePath)
	cmd := exec.Command("go", "build", "-o", output, filePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error building binary: %v", err)
	}

	logger.Debug("Saved binary to ", output)
	return output, nil
}

func run(binaryPath string) error {
	logger.Debug("run binary from : ", binaryPath)
	cmd := exec.Command(binaryPath)

	// Redirect standard input, output, and error to the current process
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running binary : %v", err)
	}

	return nil

}
