package generate

import (
	"fmt"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/spf13/cobra"
)

type Flags struct {
	ConfigPath  string
	ProjectPath string
}

func Command() *cobra.Command {
	flags := Flags{}

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate application resource",
		Long:  "Generate deployment manifest and main function backend application",
		RunE:  generateCmd(&flags),
	}

	cmd.Flags().StringVarP(&flags.ConfigPath, "config", "c", "", "Path to the configuration file")
	cmd.Flags().StringVarP(&flags.ProjectPath, "project", "p", "", "Path to project folder")

	return cmd
}

func generateCmd(flags *Flags) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// load config
		config, err := raiden.LoadConfig(&flags.ConfigPath)
		if err != nil {
			return err
		}

		if flags.ProjectPath != "" {
			config.ProjectName = flags.ProjectPath
		}

		// generate resource
		if err := GenerateResource(config); err != nil {
			return err
		}

		return nil
	}
}

func GenerateResource(config *raiden.Config) error {
	// get project id
	var projectId *string
	if config.DeploymentTarget == raiden.DeploymentTargetCloud {
		pId, err := getProjectId(config)
		if err != nil {
			return err
		}

		if pId == nil {
			return fmt.Errorf("project %s is not found in supabase cloud", config.ProjectName)
		}

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
	if err := generator.GenerateRoute("", config.ProjectName, generator.Generate); err != nil {
		return err
	}

	// TODO : synchronize local and cloud before generate

	// generate all model from cloud / pg-meta
	if err := generator.GenerateModels("", tables, policies, generator.Generate); err != nil {
		return err
	}

	// generate all roles from cloud / pg-meta
	if err := generator.GenerateRoles("", roles, generator.Generate); err != nil {
		return err
	}

	return nil
}

func getProjectId(config *raiden.Config) (*string, error) {
	var projectName string
	projectNameArr := strings.Split(strings.TrimRight(config.ProjectName, "/"), "/")
	if len(projectNameArr) > 1 {
		projectName = projectNameArr[len(projectNameArr)-1]
	} else {
		projectName = projectNameArr[0]
	}

	supabase.ConfigureManagementApi(config.SupabaseApiUrl, config.AccessToken)
	project, err := supabase.FindProject(projectName)
	if err != nil {
		return nil, err
	}

	if project.Id == "" {
		err = fmt.Errorf("%s is not exist, creating new project", config.ProjectName)
		return nil, err
	}

	return &project.Id, nil
}
