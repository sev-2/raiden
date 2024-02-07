package init

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

// The `Flags` type represents a set of command line flags, with a `Version` field for specifying the
// version of a raiden.
// @property {string} Version - The `Version` property is a string that represents the version of the
// Raiden software. It can be specified using a tag, branch, or commit number.
type Flags struct {
	Target string
}

func (f *Flags) Bind(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.Target, "target", "t", "targeted specific raiden version from tag, branch or commit number")
}

func PreRun(projectPath string) error {
	if !configure.IsConfigExist(projectPath) {
		return errors.New("missing config file (./configs/app.yaml), run `raiden configure` first for generate configuration file")
	}

	return nil
}

// The `Run` function initializes a Go module, checks if a `go.mod` file exists, prompts the user to
// re-initialize the module if it does, deletes the `go.mod` and `go.sum` files if necessary, runs `go
// mod init` with the specified module name,for gets the raiden library from a repository, and finally runs
// `go mod tidy` to install dependencies.
func Run(flags *Flags, projectPath string, moduleName string) error {
	logger.Debug("Change directory to : ", projectPath)
	if err := os.Chdir(projectPath); err != nil {
		return err
	}

	// check go.mod exist
	logger.Debug("check is module initialize")
	goModFile := filepath.Join(projectPath, "go.mod")
	if IsModFileExist(projectPath) {
		input := confirmation.New("Found go.mod file, do you want to re init module ?", confirmation.No)
		input.DefaultValue = confirmation.No

		isReInit, err := input.RunPrompt()
		if err != nil {
			return err
		}

		if !isReInit {
			return nil
		}

		utils.DeleteFile(goModFile)
		utils.DeleteFile(filepath.Join(projectPath, "go.sum"))
	}

	// running go mod init
	logger.Debug("Execute command : go mod init ", moduleName)

	cmdModInit := exec.Command("go", "mod", "init", moduleName)
	cmdModInit.Stdout = os.Stdout
	cmdModInit.Stderr = os.Stderr

	if err := cmdModInit.Run(); err != nil {
		return fmt.Errorf("error init project : %v", err)
	}

	// get raiden app
	repoName := "github.com/sev-2/raiden"
	if flags.Target != "" {
		repoName += "@" + flags.Target
	}
	logger.Debugf("Execute command : go get %s ", repoName)
	cmdRaidenInit := exec.Command("go", "get", repoName)
	cmdRaidenInit.Stdout = os.Stdout
	cmdRaidenInit.Stderr = os.Stderr

	if err := cmdRaidenInit.Run(); err != nil {
		return fmt.Errorf("error get raiden app : %v", err)
	}

	// mod tidy
	logger.Debug("Execute command : go mod tidy")
	cmdModTidy := exec.Command("go", "mod", "tidy")
	cmdModTidy.Stdout = os.Stdout
	cmdModTidy.Stderr = os.Stderr

	if err := cmdModTidy.Run(); err != nil {
		return fmt.Errorf("error install dependency : %v", err)
	}

	return nil
}

func IsModFileExist(projectPath string) bool {
	goModFile := filepath.Join(projectPath, "go.mod")
	return utils.IsFileExists(goModFile)
}
