package start

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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
		createInput.SupabasePublicUrl = fmt.Sprintf("https://%s.supabase.co/", project.Id)
	case raiden.DeploymentTargetSelfHosted:
		supabase.ConfigurationMetaApi(createInput.SupabaseApiUrl, createInput.SupabaseApiBasePath)
	}

	if err := generateResource(projectID, createInput); err != nil {
		return err
	}

	return initProject(createInput)
}

// generate new resource.
// contain supabase resource like table, roles, policy and etc
// also generate framework resource like controller, route, main function and etc
func generateResource(projectID *string, createInput *CreateInput) error {

	// get supabase resources from cloud or pg-meta
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
	appConfig := createInput.ToAppConfig()
	if err := generator.GenerateConfig(appConfig, generator.Generate); err != nil {
		return err
	}

	// generate example controller
	if err := generator.GenerateHelloWordController(createInput.ProjectName, generator.Generate); err != nil {
		return err
	}

	// generate all model from cloud / pg-meta
	if err := generator.GenerateModels(createInput.ProjectName, tables, policies, generator.Generate); err != nil {
		return err
	}

	// generate all roles from cloud / pg-meta
	if err := generator.GenerateRoles(createInput.ProjectName, roles, generator.Generate); err != nil {
		return err
	}

	// generate route base on controllers
	if err := generator.GenerateRoute(createInput.ProjectName, createInput.ProjectName, generator.Generate); err != nil {
		return err
	}

	if err := generator.GenerateMainFunction(appConfig, generator.Generate); err != nil {
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

func initProject(createInput *CreateInput) error {
	currentPath, err := utils.GetCurrentDirectory()
	if err != nil {
		return err
	}

	projectPath := filepath.Join(currentPath, createInput.ProjectName)
	if err := os.Chdir(projectPath); err != nil {
		return err
	}
	cmdModInit := exec.Command("go", "mod", "init", utils.ToGoModuleName(createInput.ProjectName))
	cmdModInit.Stdout = os.Stdout
	cmdModInit.Stderr = os.Stderr

	if err := cmdModInit.Run(); err != nil {
		return fmt.Errorf("error init project : %v", err)
	}

	cmdModTidy := exec.Command("go", "mod", "tidy")
	cmdModTidy.Stdout = os.Stdout
	cmdModTidy.Stderr = os.Stderr

	if err := cmdModTidy.Run(); err != nil {
		return fmt.Errorf("error install dependency : %v", err)
	}

	return nil
}
