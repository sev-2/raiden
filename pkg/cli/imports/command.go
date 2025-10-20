package imports

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden/pkg/cli"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

var ImportLogger hclog.Logger = logger.HcLog().Named("import")

var buildDir = "build"

type Flags struct {
	RpcOnly            bool
	RolesOnly          bool
	ModelsOnly         bool
	StoragesOnly       bool
	PolicyOnly         bool
	AllowedSchema      string
	GenerateController bool
	ForceImport        bool
	DryRun             bool
}

func (f *Flags) Bind(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&f.RpcOnly, "rpc-only", "", false, "import rpc only")
	cmd.Flags().BoolVarP(&f.RolesOnly, "roles-only", "r", false, "import roles only")
	cmd.Flags().BoolVarP(&f.ModelsOnly, "models-only", "m", false, "import models only")
	cmd.Flags().BoolVarP(&f.PolicyOnly, "policy-only", "", false, "import policy only")
	cmd.Flags().BoolVarP(&f.StoragesOnly, "storages-only", "", false, "import storage only")
	cmd.Flags().StringVarP(&f.AllowedSchema, "schema", "s", "", "set allowed schema to import, use coma separator for multiple schema")
	cmd.Flags().BoolVar(&f.GenerateController, "generate-controller", false, "generate rest controllers for imported tables")
	cmd.Flags().BoolVar(&f.ForceImport, "force", false, "skip diff checks and overwrite local resource state")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "run import in simulate mode without actual import resource as code")

}

func (f *Flags) LoadAll() bool {
	return !f.RpcOnly && !f.RolesOnly && !f.ModelsOnly && !f.PolicyOnly
}

func PreRun(projectPath string) error {
	if !configure.IsConfigExist(projectPath) {
		return errors.New("missing config file (./configs/app.yaml), run `raiden configure` first for generate configuration file")
	}

	return nil
}

func Run(logFlags *cli.LogFlags, flags *Flags, projectPath string) error {
	var generatedResources []string
	if flags.LoadAll() {
		generatedResources = append(generatedResources, "all")
	} else {
		if flags.ModelsOnly {
			generatedResources = append(generatedResources, "model")
		}

		if flags.RpcOnly {
			generatedResources = append(generatedResources, "rpc")
		}

		if flags.RolesOnly {
			generatedResources = append(generatedResources, "role")
		}

		if flags.StoragesOnly {
			generatedResources = append(generatedResources, "storage")
		}

		if flags.PolicyOnly {
			generatedResources = append(generatedResources, "policy")
		}
	}

	ImportLogger.Info("prepare import", "resource", strings.Join(generatedResources, ","))
	mainFile := "cmd/import/main.go"

	// set abs file path
	mainFilePath := filepath.Join(projectPath, mainFile)

	// set output
	output := GetBuildFilePath(projectPath, runtime.GOOS, "import")
	if utils.IsFileExists(output) {
		if err := utils.DeleteFile(output); err != nil {
			return err
		}
	}

	// Run the "go build" command
	ImportLogger.Debug("execute command", "cmd", fmt.Sprintf("go build -o %s %s", output, mainFilePath))
	cmd := exec.Command("go", "build", "-o", output, mainFilePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error building binary: %v", err)
	}

	ImportLogger.Debug("prepare exec binary", "path", output)
	var args []string
	if flags.RpcOnly {
		args = append(args, "--rpc-only")
	}

	if flags.ModelsOnly {
		args = append(args, "--models-only")
	}

	if flags.RolesOnly {
		args = append(args, "--roles-only")
	}

	if flags.PolicyOnly {
		args = append(args, "--policy-only")
	}

	if flags.StoragesOnly {
		args = append(args, "--storages-only")
	}

	if flags.AllowedSchema != "" {
		args = append(args, "--schema", flags.AllowedSchema)
	}

	if flags.GenerateController {
		args = append(args, "--generate-controller")
	}

	if flags.ForceImport {
		args = append(args, "--force")
	}

	if flags.DryRun {
		args = append(args, "--dry-run")
	}

	if logFlags.DebugMode {
		args = append(args, "--debug")
	} else if logFlags.TraceMode {
		args = append(args, "--trace")
	}

	ImportLogger.Info("start import")
	ImportLogger.Debug("exec binary", "path", output, "args", args)
	runCmd := exec.Command(output, args...)

	// Redirect standard input, output, and error to the current process
	runCmd.Stdin = os.Stdin
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr

	// Run the command
	return runCmd.Run()
}

func GetBuildFilePath(projectPath, targetOs string, fileName string) string {
	fullFilePath := filepath.Join(projectPath, buildDir, fileName)
	if targetOs == "windows" {
		fullFilePath += ".exe"
	}

	return fullFilePath
}
