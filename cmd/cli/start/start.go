package start

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase"
	management_api "github.com/sev-2/raiden/pkg/supabase/management-api"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start new app",
		Long:  "Start new project, synchronize resource and scaffold application",
		RunE:  createCmd(),
	}
}

func createCmd() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		createInput := &CreateInput{}
		if err := createInput.PromptAll(); err != nil {
			return err
		}

		utils.CreateFolder(createInput.ProjectName)
		err := create(cmd, args, createInput)
		if err != nil {
			utils.DeleteFolder(createInput.ProjectName)
			return err
		}

		return nil
	}
}

func create(cmd *cobra.Command, args []string, createInput *CreateInput) error {
	var projectID *string

	switch createInput.Target {
	case raiden.DeploymentTargetCloud:
		// setup management api
		supabase.ConfigureManagementApi(createInput.SupabaseApiUrl, createInput.AccessToken)
		project, err := supabase.FindProject(createInput.ProjectName)
		if err != nil {
			return err
		}

		if project.Id == "" {
			logger.Infof("%s is not exist, creating new project", createInput.ProjectName)
			project = createNewSupabaseProject(createInput.ProjectName)
		}
		projectID = &project.Id
	case raiden.DeploymentTargetSelfHosted:
		supabase.ConfigurationMetaApi(createInput.SupabaseApiUrl, createInput.SupabaseApiBasePath)
	}

	return generateResource(projectID, createInput)
}

func generateResource(projectID *string, createInput *CreateInput) error {

	// get supabase resources
	tables, err := supabase.GetTables(projectID)
	if err != nil {
		return err
	}

	roles, err := supabase.GetRoles(projectID)
	if err != nil {
		return err
	}

	policies, err := supabase.GetPolicies(projectID)
	if err != nil {
		return err
	}

	// generate resources to raiden resources
	if err := generator.GenerateConfig(*createInput.ToAppConfig()); err != nil {
		return err
	}

	// todo : generate controller

	if err := generator.GenerateModels(createInput.ProjectName, tables, policies); err != nil {
		return err
	}

	if err := generator.GenerateRoles(createInput.ProjectName, roles); err != nil {
		return err
	}

	return nil
}

func createNewSupabaseProject(projectName string) supabase.Project {
	createProjectInput := createProjectInput{
		management_api.CreateProjectBody{
			Name:       projectName,
			Plan:       "free",
			KpsEnabled: false,
		},
	}

	if err := createProjectInput.PromptAll(); err != nil {
		logger.Panic(err)
	}

	project, err := supabase.CreateNewProject(createProjectInput.CreateProjectBody)
	if err != nil {
		logger.Panic(err)
	}

	return project
}
