package start

import (
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase"
	management_api "github.com/sev-2/raiden/pkg/supabase/management-api"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start new app",
		Long:  "Start new project, synchronize resource and scaffold application",
		RunE:  createCmd,
	}
}

func createCmd(cmd *cobra.Command, args []string) error {
	createInput := &CreateInput{}
	if err := createInput.PromptAll(); err != nil {
		return err
	}

	switch createInput.Target {
	case DeploymentTargetCloud:
		cloudConfiguration(createInput)
	case DeploymentTargetSelfHosted:

	}

	// todo : generate controller

	// generate config
	return nil
}

func cloudConfiguration(createInput *CreateInput) {
	supabase.ConfigureManagementApi(createInput.SupabaseApiUrl, createInput.AccessToken)
	logger.Info("checking if project exist in supabase cloud")

	project, err := supabase.FindProject(createInput.ProjectName)
	if err != nil {
		logger.Panic(err)
	}

	if project.Id == "" {
		logger.Infof("%s is not exist, creating new project", createInput.ProjectName)
		project = createNewProject(createInput.ProjectName)
	}

	// generate existing resource (table, role, rls and etc...)
	tables, err := supabase.GetTables(&project.Id)
	if err != nil {
		logger.Panic(err)
	}
	logger.PrintJson(tables, false)

	// todo : get rls

	// todo : generate table and rls to model

	// todo : get & generate role
}

func createNewProject(projectName string) supabase.Project {
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
