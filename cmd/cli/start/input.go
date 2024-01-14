package start

import (
	"errors"

	"github.com/manifoldco/promptui"
	"github.com/sev-2/raiden/pkg/supabase"
	management_api "github.com/sev-2/raiden/pkg/supabase/management-api"
)

type DeploymentTarget string

const (
	DeploymentTargetCloud      DeploymentTarget = "cloud"
	DeploymentTargetSelfHosted DeploymentTarget = "self_hosted"
)

type CreateInput struct {
	ProjectName    string
	Target         DeploymentTarget
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
	case DeploymentTargetCloud:
		if err := ca.PromptSupabaseApiUrl(); err != nil {
			return err
		}
		if err := ca.PromptAccessToken(); err != nil {
			return err
		}
	case DeploymentTargetSelfHosted:
	}

	return nil
}

func (ca *CreateInput) PromptProjectName() error {
	promp := promptui.Prompt{
		Label: "Enter app name",
		Validate: func(s string) error {
			if s == "" {
				return errors.New("app name")
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
		Items: []DeploymentTarget{DeploymentTargetCloud, DeploymentTargetSelfHosted},
	}

	_, inputText, err := promp.Run()
	if err != nil {
		return err
	}

	ca.Target = DeploymentTarget(inputText)
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
		HideEntered: true,
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
		Default: "https://api.supabase.com",
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
