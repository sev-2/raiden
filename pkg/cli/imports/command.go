package imports

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

var buildDir = "build"

type Flags struct {
	RpcOnly       bool
	RolesOnly     bool
	ModelsOnly    bool
	AllowedSchema string
}

type MapTable map[string]*objects.Table
type MapRelations map[string][]*generator.Relation
type ManyToManyTable struct {
	Table      string
	Schema     string
	PivotTable string
	PrimaryKey string
	ForeignKey string
}

func (f *Flags) Bind(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&f.RpcOnly, "rpc-only", "", false, "import rpc only")
	cmd.Flags().BoolVarP(&f.RolesOnly, "roles-only", "r", false, "import roles only")
	cmd.Flags().BoolVarP(&f.ModelsOnly, "models-only", "m", false, "import models only")
	cmd.Flags().StringVarP(&f.AllowedSchema, "schema", "s", "", "set allowed schema to import, use coma separator for multiple schema")
}

func (f *Flags) LoadAll() bool {
	return !f.RpcOnly && !f.RolesOnly && !f.ModelsOnly
}

func PreRun(projectPath string) error {
	if !configure.IsConfigExist(projectPath) {
		return errors.New("missing config file (./configs/app.yaml), run `raiden configure` first for generate configuration file")
	}

	return nil
}

func Run(flags *Flags, projectPath string, verbose bool) error {
	mainFile := "cmd/import/main.go"

	// set abs file path
	mainFilePath := filepath.Join(projectPath, mainFile)

	// set output
	output := GetBuildFilePath(projectPath, runtime.GOOS, "import")

	if utils.IsFileExists(output) {
		utils.DeleteFile(output)
	}

	// Run the "go build" command
	logger.Debugf("Execute command : go build -o %s %s", output, mainFilePath)
	cmd := exec.Command("go", "build", "-o", output, mainFilePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error building binary: %v", err)
	}

	logger.Debug("prepare exec binary from : ", output)

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

	if flags.AllowedSchema != "" {
		args = append(args, "--schema "+flags.AllowedSchema)
	}

	if verbose {
		args = append(args, "-v")
	}

	logger.Debugf("%s %s", output, strings.Join(args, " "))
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
