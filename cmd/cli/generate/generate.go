package generate

import (
	"fmt"
	"path/filepath"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

type Flags struct {
	ConfigPath  string
	ProjectPath string
	Verbose     bool
}

func Command() *cobra.Command {
	flags := Flags{}

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate application resource",
		Long:  "Generate deployment manifest and application resource",
		RunE:  generateCmd(&flags),
	}

	cmd.Flags().StringVarP(&flags.ConfigPath, "config", "c", "", "Path to the configuration file")
	cmd.Flags().StringVarP(&flags.ProjectPath, "project", "p", "", "Path to project folder")
	cmd.PersistentFlags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Enable verbose mode")

	return cmd
}

func generateCmd(flags *Flags) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error { // set default value
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

		// generate resource
		if err := GenerateResource(flags.ProjectPath, config); err != nil {
			return err
		}

		return nil
	}
}

func GenerateResource(basePath string, config *raiden.Config) error {
	// get project id
	var projectId *string
	logger.Debug("Detect deployment target to : ", config.DeploymentTarget)
	if config.DeploymentTarget == raiden.DeploymentTargetCloud {
		logger.Debug("Get detail project : ", config.ProjectName)
		pId, err := getProjectId(config)
		if err != nil {
			return err
		}

		if pId == nil {
			return fmt.Errorf("project %s is not found in supabase cloud", config.ProjectName)
		}

		logger.Debugf("Found project id for (%s) : %s", config.ProjectName, *pId)
		projectId = pId
	}

	// get supabase resources from cloud or pg-meta
	tables, err := supabase.GetTables(projectId)
	if err != nil {
		return err
	}

	roles, err := supabase.GetRoles(projectId)
	if err != nil {
		return err
	}

	policies, err := supabase.GetPolicies(projectId)
	if err != nil {
		return err
	}

	// generate route base on controllers
	if err := generator.GenerateRoute(basePath, config.ProjectName, generator.Generate); err != nil {
		return err
	}

	// TODO : synchronize local and cloud before generate

	// generate all model from cloud / pg-meta
	if err := generator.GenerateModels(basePath, tables, policies, generator.Generate); err != nil {
		return err
	}

	// generate all roles from cloud / pg-meta
	if err := generator.GenerateRoles(basePath, roles, generator.Generate); err != nil {
		return err
	}

	return nil
}

func getProjectId(config *raiden.Config) (*string, error) {
	supabase.ConfigureManagementApi(config.SupabaseApiUrl, config.AccessToken)
	project, err := supabase.FindProject(config.ProjectName)
	if err != nil {
		return nil, err
	}

	if project.Id == "" {
		err = fmt.Errorf("%s is not exist, creating new project", config.ProjectName)
		return nil, err
	}

	return &project.Id, nil
}
