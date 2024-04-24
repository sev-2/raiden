package serve

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli/build"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
)

var ServeLogger hclog.Logger = logger.HcLog().Named("serve")

// The function `PreRun` checks if the necessary configuration and initialization files exist in the
// specified project path.
func PreRun(projectPath string) error {
	if !configure.IsConfigExist(projectPath) {
		return errors.New("missing config file (./configs/app.yaml), run `raiden configure` first for generate configuration file")
	}
	return nil
}

func Run(config *raiden.Config, projectPath string) error {
	ServeLogger.Info("prepare configuration")
	binaryPath := build.GetBuildFilePath(projectPath, config.ProjectName, runtime.GOOS)
	ServeLogger.Debug("check app binary", "path", binaryPath)
	if !utils.IsFileExists(binaryPath) {
		return errors.New("app binary file not found, run `raiden build` first for build app binary")
	}

	ServeLogger.Debug("run binary", "path", binaryPath)
	cmd := exec.Command(binaryPath)

	// Redirect standard input, output, and error to the current process
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	ServeLogger.Info("start running server")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running binary : %v", err)
	}

	return nil
}
