package start

import (
	"errors"

	"github.com/manifoldco/promptui"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase"
	management_api "github.com/sev-2/raiden/pkg/supabase/management-api"
	"github.com/sev-2/raiden/pkg/utils"
)

type CreateInput struct {
	ProjectName    string
	Target         raiden.DeploymentTarget
	AccessToken    string
	SupabaseApiUrl string
}

func (ca *CreateInput) PromptAll() error {
	if err := ca.PromptProjectName(); err != nil {
		return err
	}

	if err := ca.PromptTarget(); err != nil {
		return err
	}

	switch ca.Target {
	case raiden.DeploymentTargetCloud:
		if err := ca.PromptSupabaseApiUrl(); err != nil {
			return err
		}
		if err := ca.PromptAccessToken(); err != nil {
			return err
		}
	case raiden.DeploymentTargetSelfHosted:
	}

	return nil
}

func (ca *CreateInput) ToAppConfig() *raiden.Config {
	return &raiden.Config{
		ProjectName:      ca.ProjectName,
		DeploymentTarget: ca.Target,
		CloudAccessToken: ca.AccessToken,
		SupabaseApiUrl:   ca.SupabaseApiUrl,
	}
}

func (ca *CreateInput) PromptProjectName() error {
	promp := promptui.Prompt{
		Label: "Enter project name",
		Validate: func(s string) error {
			if s == "" {
				return errors.New("project name can`t be empty")
			}

			if utils.IsStringContainSpace(s) {
				return errors.New("project name can`t contain spaces character")
			}

			return nil
		},
	}

	inputText, err := promp.Run()
	if err != nil {
		return err
	}

	ca.ProjectName = inputText
	return nil
}

func (ca *CreateInput) PromptTarget() error {
	promp := promptui.Select{
		Label: "Enter your target deployment",
		Items: []raiden.DeploymentTarget{raiden.DeploymentTargetCloud, raiden.DeploymentTargetSelfHosted},
	}

	_, inputText, err := promp.Run()
	if err != nil {
		return err
	}

	ca.Target = raiden.DeploymentTarget(inputText)
	return nil
}

func (ca *CreateInput) PromptAccessToken() error {
	promp := promptui.Prompt{
		Label: "Enter your access token",
		Validate: func(s string) error {
			if s == "" {
				return errors.New("access token can`t be empty")
			}

			return nil
		},
		Mask: '*',
	}

	inputText, err := promp.Run()
	if err != nil {
		return err
	}

	ca.AccessToken = inputText
	return nil
}

func (ca *CreateInput) PromptSupabaseApiUrl() error {
	promp := promptui.Prompt{
		Label:   "Enter supabase api url",
		Default: supabase.DefaultApiUrl,
	}

	inputText, err := promp.Run()
	if err != nil {
		return err
	}

	ca.SupabaseApiUrl = inputText
	return nil
}

// create project input
type createProjectInput struct {
	management_api.CreateProjectBody
}

func (ca *createProjectInput) PromptAll() error {
	if err := ca.PromptOrganizations(); err != nil {
		return err
	}

	if err := ca.PromptRegion(); err != nil {
		return err
	}

	if err := ca.PromptDbPassword(); err != nil {
		return err
	}

	return nil
}

func (cp *createProjectInput) PromptOrganizations() error {
	organizations, err := supabase.GetOrganizations()
	if err != nil {
		return err
	}

	// todo : create new organization if doesn`t have organization
	var organizationOptions []string
	var selectedOrganization string

	for _, org := range organizations {
		organizationOptions = append(organizationOptions, org.Name)
	}

	promp := promptui.Select{
		Label: "Select organization",
		Items: organizationOptions,
	}

	_, inputText, err := promp.Run()
	if err != nil {
		return err
	}

	for _, org := range organizations {
		if org.Name == inputText {
			selectedOrganization = org.Id
		}
	}

	cp.OrganizationId = selectedOrganization
	return nil
}

func (cp *createProjectInput) PromptRegion() error {
	promp := promptui.Select{
		Label: "Select regions",
		Items: supabase.AllRegion,
	}

	_, inputText, err := promp.Run()
	if err != nil {
		return err
	}

	cp.Region = inputText
	return nil
}

func (cp *createProjectInput) PromptDbPassword() error {
	promp := promptui.Prompt{
		Label: "Set database password",
		Validate: func(s string) error {
			if s == "" {
				return errors.New("database password can`t be empty")
			}
			return nil
		},
		Mask:        '*',
		HideEntered: true,
	}

	inputText, err := promp.Run()
	if err != nil {
		return err
	}

	cp.DbPass = inputText
	return nil
}
