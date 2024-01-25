package start

import (
	"errors"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase"
	management_api "github.com/sev-2/raiden/pkg/supabase/management-api"
	"github.com/sev-2/raiden/pkg/utils"
)

type CreateInput struct {
	AccessToken         string
	ProjectName         string
	GoModuleName        string
	SupabaseApiUrl      string
	SupabaseApiBasePath string
	Target              raiden.DeploymentTarget
	TraceEnable         bool
	TraceEndpoint       string
	TraceCollector      string
}

func (ca *CreateInput) PromptAll() error {
	if err := ca.PromptProjectName(); err != nil {
		return err
	}

	if err := ca.PromptGoModuleName(); err != nil {
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
		if err := ca.PromptSupabaseApiUrl(); err != nil {
			return err
		}

		if err := ca.PromptSupabaseApiPath(); err != nil {
			return err
		}
	}

	if err := ca.PrompTraceEnable(); err != nil {
		return err
	}

	if ca.TraceEnable {
		if err := ca.PromptTraceCollector(); err != nil {
			return err
		}

		if err := ca.PromptTraceEndpoint(); err != nil {
			return err
		}
	}
	return nil
}

func (ca *CreateInput) ToAppConfig() *raiden.Config {
	return &raiden.Config{
		CloudAccessToken:   ca.AccessToken,
		DeploymentTarget:   ca.Target,
		ProjectName:        ca.ProjectName,
		GoModuleName:       ca.GoModuleName,
		SupabaseApiUrl:     ca.SupabaseApiUrl,
		SupabaseApiBaseUrl: ca.SupabaseApiBasePath,
		TraceEnable:        ca.TraceEnable,
		TraceCollector:     ca.TraceCollector,
		TraceEndpoint:      ca.TraceEndpoint,
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

func (ca *CreateInput) PromptGoModuleName() error {
	promp := promptui.Prompt{
		Label:   "Enter go module name",
		Default: utils.ToGoModuleName(ca.ProjectName),
	}

	inputText, err := promp.Run()
	if err != nil {
		return err
	}

	ca.GoModuleName = inputText
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
	defaultValue := "http://localhost:54323"
	switch ca.Target {
	case raiden.DeploymentTargetCloud:
		defaultValue = supabase.DefaultApiUrl
		// other target
	}

	promp := promptui.Prompt{
		Label:   "Enter supabase api url",
		Default: defaultValue,
	}

	inputText, err := promp.Run()
	if err != nil {
		return err
	}

	ca.SupabaseApiUrl = inputText
	return nil
}

func (ca *CreateInput) PromptSupabaseApiPath() error {
	promp := promptui.Prompt{
		Label:   "Enter supabase api base path",
		Default: "/api/pg-meta/default",
	}

	inputText, err := promp.Run()
	if err != nil {
		return err
	}

	ca.SupabaseApiBasePath = inputText
	return nil
}

func (ca *CreateInput) PrompTraceEnable() error {
	promp := promptui.Prompt{
		Label:     "Enable Tracer",
		Default:   "n",
		IsConfirm: true,
	}

	inputText, err := promp.Run()
	if err != nil {
		return err
	}

	if strings.ToLower(inputText) == "y" {
		ca.TraceEnable = true
	} else {
		ca.TraceEnable = false
	}
	return nil
}

func (ca *CreateInput) PromptTraceCollector() error {
	promp := promptui.Select{
		Label: "Choose collector",
		Items: []string{"otpl"},
	}

	_, inputText, err := promp.Run()
	if err != nil {
		return err
	}
	ca.TraceCollector = inputText
	return nil
}

func (ca *CreateInput) PromptTraceEndpoint() error {
	promp := promptui.Prompt{
		Label:   "Enter trace endpoint",
		Default: "localhost:4317",
		Validate: func(s string) error {
			if s == "" {
				return errors.New("trace endpoint can`t be empty")
			}
			return nil
		},
	}

	inputText, err := promp.Run()
	if err != nil {
		return err
	}
	ca.TraceEndpoint = inputText
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
