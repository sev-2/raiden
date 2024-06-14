package generate

import (
	"errors"
	"strings"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/cli/configure"
	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/spf13/cobra"
)

var GenerateLogger hclog.Logger = logger.HcLog().Named("generate")

// The above type represents a set of flags used for configuring different modes in a Go program.
// @property {bool} RpcOnly - RpcOnly is a boolean flag that indicates whether the program should only
// run in RPC mode.
// @property {bool} RoutesOnly - RoutesOnly is a boolean flag that indicates whether only the routes
// should be processed.
// @property {bool} InitMode - The `InitMode` property is a boolean flag that indicates whether the
// program should run in initialization mode.
type Flags struct {
	RpcOnly    bool
	RoutesOnly bool
	InitMode   bool
}

// The `Bind` function is used to bind the flags of the `Flags` struct to a command in the Cobra
// library. It takes a `cmd` parameter, which is a command in the Cobra library, and calls the
// `BindRouteOnly` and `BindRpcOnly` functions to bind the corresponding flags to the command.
func (f *Flags) Bind(cmd *cobra.Command) {
	f.BindRouteOnly(cmd)
	f.BindRpcOnly(cmd)
}

// The `BindRpcOnly` function is used to bind the `RpcOnly` flag to a command in the Cobra library. It
// checks if the command already has a flag named "rpc-only" and if so, it binds the `RpcOnly` flag to
// that flag. If the command does not have a flag named "rpc-only", it creates a new flag with the name
// "rpc-only" and binds the `RpcOnly` flag to it. The function also sets the default value and usage
// message for the flag.
func (f *Flags) BindRpcOnly(cmd *cobra.Command) {
	if cmd.Flags().Lookup("rpc-only") != nil {
		cmd.Flags().BoolVar(&f.RpcOnly, "generate-rpc-only", false, "generate register rpc file only")
	} else {
		cmd.Flags().BoolVar(&f.RpcOnly, "rpc-only", false, "generate register rpc file only")
	}
}

// The `BindRouteOnly` function is used to bind the `RoutesOnly` flag to a command in the Cobra
// library.
func (f *Flags) BindRouteOnly(cmd *cobra.Command) {
	shortHand := "r"

	if cmd.Flags().ShorthandLookup("r") != nil {
		shortHand = ""
	}

	if cmd.Flags().Lookup("rpc-only") != nil {
		cmd.Flags().BoolVarP(&f.RoutesOnly, "generate-routes-only", shortHand, false, "generate register routes file only")
	} else {
		cmd.Flags().BoolVarP(&f.RoutesOnly, "routes-only", shortHand, false, "generate register routes file only")
	}
}

func (f *Flags) IsGenerateAll() bool {
	return !f.RpcOnly && !f.RoutesOnly
}

func PreRun(projectPath string) error {
	if !configure.IsConfigExist(projectPath) {
		return errors.New("missing config file (./configs/app.yaml), run `raiden configure` first for generate configuration file")
	}

	return nil
}

// The `Run` function generates code based on the provided flags and configuration, including
// generating controllers, routes, and a main function.
func Run(flags *Flags, config *raiden.Config, projectPath string, initialize bool) error {
	var generatedModule []string
	if flags.IsGenerateAll() {
		generatedModule = append(generatedModule, "all")
	} else {
		if flags.RoutesOnly {
			generatedModule = append(generatedModule, "route")
		}

		if flags.RpcOnly {
			generatedModule = append(generatedModule, "rpc")
		}
	}

	GenerateLogger.Debug("prepare generate file", "modules", strings.Join(generatedModule, ","))
	if err := generator.CreateInternalFolder(projectPath); err != nil {
		return err
	}

	wg, errChan := sync.WaitGroup{}, make(chan error)
	go func() {
		wg.Wait()
		close(errChan)
	}()

	if flags.IsGenerateAll() || flags.RoutesOnly {
		wg.Add(1)
		go func() {
			defer wg.Done()
			GenerateLogger.Debug("start generate routes")
			if initialize {
				// generate example controller
				GenerateLogger.Info("start generate hello world controller")
				if err := generator.GenerateHelloWorldController(projectPath, generator.Generate); err != nil {
					errChan <- err
					return
				}
				GenerateLogger.Info("finish generate hello world controller")
			}

			// generate route base on controllers
			if err := generator.GenerateRoute(projectPath, config.ProjectName, generator.Generate); err != nil {
				errChan <- err
				return
			}
			errChan <- nil
			GenerateLogger.Info("finish generate routes")
		}()
	}

	if initialize {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// generate main function
			GenerateLogger.Info("start generate main function")
			if err := generator.GenerateMainFunction(projectPath, config, generator.Generate); err != nil {
				errChan <- err
			} else {
				errChan <- nil
			}
			GenerateLogger.Info("finish generate main function")
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		// generate rpc register
		GenerateLogger.Debug("start generate rpc register file")
		if err := generator.GenerateRpcRegister(projectPath, config.ProjectName, generator.Generate); err != nil {
			errChan <- err
		}
		GenerateLogger.Debug("finish generate rpc register file")

		// generate role register
		GenerateLogger.Debug("start generate role register file")
		if err := generator.GenerateRoleRegister(projectPath, config.ProjectName, generator.Generate); err != nil {
			errChan <- err
		}
		GenerateLogger.Debug("finish generate role register file")

		// generate model register
		GenerateLogger.Debug("start generate model register file")
		if err := generator.GenerateModelRegister(projectPath, config.ProjectName, generator.Generate); err != nil {
			errChan <- err
		}
		GenerateLogger.Debug("finish generate role register file")

		// generate storage register
		GenerateLogger.Debug("start generate storages register file")
		if err := generator.GenerateStoragesRegister(projectPath, config.ProjectName, generator.Generate); err != nil {
			errChan <- err
		}
		GenerateLogger.Debug("finish generate storages register file")

		// generate job register
		GenerateLogger.Debug("start generate job register file")
		if err := generator.GenerateJobRegister(projectPath, config.ProjectName, generator.Generate); err != nil {
			errChan <- err
		}
		GenerateLogger.Debug("finish generate job register file")

		if initialize {
			// generate import main function
			GenerateLogger.Debug("start generate import main function file")
			if err := generator.GenerateImportMainFunction(projectPath, config, generator.Generate); err != nil {
				errChan <- err
			}
			GenerateLogger.Debug("finish generate import main function file")

			// generate import main function
			GenerateLogger.Debug("start generate apply main function file")
			if err := generator.GenerateApplyMainFunction(projectPath, config, generator.Generate); err != nil {
				errChan <- err
			} else {
				errChan <- nil
			}
			GenerateLogger.Debug("finish generate import main function file")
		} else {
			errChan <- nil
		}
	}()

	var err error
	for rsErr := range errChan {
		if rsErr != nil && err == nil {
			err = rsErr
		}
	}

	return err
}
