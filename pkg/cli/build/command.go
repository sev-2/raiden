package build

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli/configure"
	init_command "github.com/sev-2/raiden/pkg/cli/init"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

var BuildLogger hclog.Logger = logger.HcLog().Named("build")

// The `type Flags struct` is defining a struct called `Flags`. This struct is used to store the values
// of command-line flags related to the target operating system and processor architecture.
type Flags struct {
	OS   string
	Arch string
}

var buildDir = "build"

// The `Bind` method is used to bind the `Flags` struct to a `cobra.Command` object. It
// sets up the command-line flags for the `OS` and `Arch` fields of the `Flags` struct,
// allowing the user to specify the target operating system and processor architecture
// when running the command.
func (f *Flags) Bind(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.OS, "os", "", "specifiy target operating system")
	cmd.Flags().StringVar(&f.Arch, "arch", "", "specifiy target processor architecture")
}

// The function `PreRun` checks if the necessary configuration and initialization files exist in the
// specified project path.
func PreRun(projectPath string) error {
	if !configure.IsConfigExist(projectPath) {
		return errors.New("missing config file (./configs/app.yaml), run `raiden configure` first for generate configuration file")
	}

	if !init_command.IsModFileExist(projectPath) {
		return errors.New("missing go.mod file, run `raiden init` first for init project")
	}

	return nil
}

// The Run function builds a Go binary based on the provided flags, configuration, and project path.
func Run(f *Flags, config *raiden.Config, projectPath string) error {
	BuildLogger.Info("start build binary")
	// determine the target operating system and processor architecture for thebuild process.
	targetOs, targetArch := runtime.GOOS, runtime.GOARCH
	if f.OS != "" {
		targetOs = f.OS
	}

	if f.Arch != "" {
		targetArch = f.Arch
	}
	// Set environment variables
	os.Setenv("GOOS", targetOs)
	os.Setenv("GOARCH", targetArch)

	// define file path for the main Go file of the project.
	mainFileName := GetBuildFileName(config.ProjectName)
	mainFile := fmt.Sprintf("cmd/%s/%s.go", config.ProjectName, mainFileName)

	// set abs file path
	mainFilePath := filepath.Join(projectPath, mainFile)

	// set output
	output := GetBuildFilePath(projectPath, config.ProjectName, targetOs)

	// Run the "go build" command
	BuildLogger.Debug("execute command", "cmd", fmt.Sprintf("go build -o %s %s", output, mainFilePath))
	cmd := exec.Command("go", "build", "-o", output, mainFilePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error building binary: %v", err)
	}

	BuildLogger.Debug("saved binary to ", "path", output)
	return nil
}

func GetBuildFileName(projectName string) string {
	return utils.ToSnakeCase(projectName)
}

func GetBuildFilePath(projectPath, projectName string, targetOs string) string {
	fullFilePath := filepath.Join(projectPath, buildDir, GetBuildFileName(projectName))
	if targetOs == "windows" {
		fullFilePath += ".exe"
	}

	return fullFilePath
}

func IsBuildFileExist(projectPath string, projectName string, targetOs string) bool {
	return utils.IsFileExists((GetBuildFilePath(projectPath, projectName, targetOs)))
}
