package configure

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/supabase"
	management_api "github.com/sev-2/raiden/pkg/supabase/management-api"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/spf13/cobra"
)

type (
	// The `Config` struct is embedding the `raiden.Config` struct. This means that the `Config` struct
	// inherits all the fields and methods of the `raiden.Config` struct, making it easier to work with
	// the configuration data. It allows the `Config` struct to have access to all the fields and methods
	// defined in the `raiden.Config` struct without having to redefine them.
	Config struct {
		raiden.Config
	}

	// The `CreateNewProjectConfig` struct is used to store the configuration data for creating a new
	// project. It embeds the `management_api.CreateProjectBody` struct, which contains the fields
	// required to create a project in the Supabase cloud. By embedding the `CreateProjectBody` struct,
	// the `CreateNewProjectConfig` struct inherits all the fields and methods of the `CreateProjectBody`
	// struct, making it easier to work with the configuration data.
	CreateNewProjectConfig struct {
		management_api.CreateProjectBody
	}

	// The `Flags` struct is used to store the command line flags that can be passed to the program. In
	// this case, it has a single boolean field `Advance` which represents whether the program should run
	// in advance mode or not. This flag is used to determine the flow of the configuration process.
	Flags struct {
		Advance bool
	}
)

// The `Bind` method is used to bind the `Flags` struct to the command line flags of a `cobra.Command`
// object. It takes a `cobra.Command` object as a parameter and sets up a boolean flag named "advance"
// with a default value of `false`. The value of this flag will be stored in the `Advance` field of the
// `Flags` struct. This allows the program to read the value of the "advance" flag from the command
// line and use it to determine whether to run in advance mode or not.
func (f *Flags) Bind(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&f.Advance, "advance", false, "configure with advance mode")
}

// The Run function checks if a config file exists and prompts the user to reconfigure if desired, then
// either performs a simple or advanced configuration based on the provided flags.
func Run(f *Flags, projectPath string) (*Config, error) {
	if IsConfigExist(projectPath) {
		input := confirmation.New("found config file, do you want to reconfigure ?", confirmation.Undecided)
		input.DefaultValue = confirmation.No

		isReconfigure, err := input.RunPrompt()
		if err != nil {
			return nil, err
		}

		if !isReconfigure {
			return nil, nil
		}
	}

	if f.Advance {
		return AdvanceConfigure()
	}
	return SimpleConfigure()
}

// The function `SimpleConfigure` prompts the user for configuration options and returns a `Config`
// object or an error.
func SimpleConfigure() (*Config, error) {
	config := &Config{
		raiden.Config{
			BreakerEnable: true,
			TraceEnable:   false,
			Version:       "1.0.0",
			ServerHost:    "127.0.01",
			ServerPort:    "8002",
		},
	}

	if err := PromptTargetDeployment(config); err != nil {
		return nil, err
	}

	if err := PromptProjectName(config); err != nil {
		return nil, err
	}

	// Prompt for cloud deployment
	if config.DeploymentTarget == raiden.DeploymentTargetCloud {
		if err := PromptAccessToken(config); err != nil {
			return nil, err
		}

		// set default cloud supabase api url
		config.SupabaseApiUrl = supabase.DefaultApiUrl

		isProjectExist, err := BindProjectIpAndPublicUrl(config)
		if err != nil {
			return nil, err
		}

		if !isProjectExist {
			return nil, fmt.Errorf("project %s is not exist, please create project first in supabase cloud", config.ProjectName)
		}
	}

	// Prompt for self host deployment
	if config.DeploymentTarget == raiden.DeploymentTargetSelfHosted {
		if err := PromptSupabaseApiUrl(config); err != nil {
			return nil, err
		}

		if err := PromptSupabaseApiPath(config); err != nil {
			return nil, err
		}

		if err := PromptSupabasePublicUrl(config); err != nil {
			return nil, err
		}
	}

	// Prompt Key
	if err := PromptAnonKey(config); err != nil {
		return nil, err
	}

	if err := PromptServiceKey(config); err != nil {
		return nil, err
	}

	return config, nil
}

func AdvanceConfigure() (*Config, error) {
	config, err := SimpleConfigure()
	if err != nil {
		return nil, err
	}

	if config == nil {
		return nil, errors.New("something wrong when create configuration, try again")
	}

	// Prompt Trace
	if err := PromptTraceEnable(config); err != nil {
		return nil, err
	}

	if config.TraceEnable {
		if err := PromptTraceCollector(config); err != nil {
			return nil, err
		}

		if err := PromptTraceCollectorEndpoint(config); err != nil {
			return nil, err
		}
	}

	// Prompt Breaker
	if err := PromptBreakerEnable(config); err != nil {
		return nil, err
	}

	return config, nil
}

func PromptTargetDeployment(c *Config) error {
	input := selection.New("Enter your target deployment", []raiden.DeploymentTarget{raiden.DeploymentTargetCloud, raiden.DeploymentTargetSelfHosted})
	input.PageSize = 2

	inputText, err := input.RunPrompt()
	if err != nil {
		return err
	}
	c.DeploymentTarget = raiden.DeploymentTarget(inputText)
	return nil
}

func PromptProjectName(c *Config) error {
	input := textinput.New("Enter project name")
	input.Validate = func(s string) error {
		if s == "" {
			return errors.New("project name can`t be empty")
		}

		if utils.IsStringContainSpace(s) {
			return errors.New("project name can`t contain spaces character")
		}

		return nil
	}

	inputText, err := input.RunPrompt()
	if err != nil {
		return err
	}

	c.ProjectName = inputText
	return nil
}

// ----- Prompt Cloud Deployment -----

func PromptAccessToken(c *Config) error {
	input := textinput.New("Enter your access token")
	input.Validate = func(s string) error {
		if s == "" {
			return errors.New("access token can`t be empty")
		}

		return nil
	}
	input.Hidden = true
	input.HideMask = '*'

	inputText, err := input.RunPrompt()
	if err != nil {
		return err
	}

	c.AccessToken = inputText
	return nil
}

func PromptCreateNewProjectConfirmation(c *Config) (isCreateNew bool, err error) {
	label := fmt.Sprintf("Project %s is not exist in supabase cloud, do you want create new ?", c.ProjectName)
	input := confirmation.New(label, confirmation.Yes)
	input.DefaultValue = confirmation.Yes
	return input.RunPrompt()
}

// ----- Prompt Key -----

func PromptAnonKey(c *Config) error {
	input := textinput.New("Enter your anon key")
	input.Validate = func(s string) error {
		if s == "" {
			return errors.New("anon key can`t be empty")
		}

		return nil
	}
	input.Hidden = true
	input.HideMask = '*'

	inputText, err := input.RunPrompt()
	if err != nil {
		return err
	}

	c.AnonKey = inputText
	return nil
}

func PromptServiceKey(c *Config) error {
	input := textinput.New("Enter your service key")
	input.Validate = func(s string) error {
		if s == "" {
			return errors.New("service key can`t be empty")
		}
		return nil
	}
	input.Hidden = true
	input.HideMask = '*'

	inputText, err := input.RunPrompt()
	if err != nil {
		return err
	}

	c.ServiceKey = inputText
	return nil
}

// ----- Prompt Self Hosted Deployment ----

func PromptSupabaseApiUrl(c *Config) error {
	input := textinput.New("Enter supabase api url")
	input.InitialValue = "http://localhost:54323"

	inputText, err := input.RunPrompt()
	if err != nil {
		return err
	}

	c.SupabaseApiUrl = inputText
	return nil
}

func PromptSupabaseApiPath(c *Config) error {
	input := textinput.New("Enter supabase api base path")
	input.InitialValue = "/api/pg-meta/default"

	inputText, err := input.RunPrompt()
	if err != nil {
		return err
	}

	c.SupabaseApiBaseUrl = inputText
	return nil
}

func PromptSupabasePublicUrl(c *Config) error {
	input := textinput.New("Enter supabase public url")

	inputText, err := input.RunPrompt()
	if err != nil {
		return err
	}
	c.SupabasePublicUrl = inputText
	return nil
}

// ----- Prompt Trace -----

func PromptTraceEnable(c *Config) error {
	input := confirmation.New("Enable Tracer ?", confirmation.Undecided)
	input.DefaultValue = confirmation.No

	inputBool, err := input.RunPrompt()
	if err != nil {
		return err
	}

	c.TraceEnable = inputBool
	return nil
}

func PromptTraceCollector(c *Config) error {
	input := selection.New("Choose collector", []string{"otpl"})
	input.PageSize = 4

	inputText, err := input.RunPrompt()
	if err != nil {
		return err
	}
	c.TraceCollector = inputText
	return nil
}

func PromptTraceCollectorEndpoint(c *Config) error {
	input := textinput.New("Enter collector endpoint")
	input.InitialValue = "localhost:4317"
	input.Validate = func(s string) error {
		if s == "" {
			return errors.New("trace collector can`t be empty")
		}
		return nil
	}

	inputText, err := input.RunPrompt()
	if err != nil {
		return err
	}
	c.TraceCollectorEndpoint = inputText
	return nil
}

// ----- Prompt Breaker -----

func PromptBreakerEnable(c *Config) error {
	input := confirmation.New("Enable Circuit Breaker ?", confirmation.Yes)
	input.DefaultValue = confirmation.Yes

	inputBool, err := input.RunPrompt()
	if err != nil {
		return err
	}

	c.BreakerEnable = inputBool
	return nil
}

// ----- Create New Project -----

func PromptCreateNewProjectConfig(c *Config) (*CreateNewProjectConfig, error) {
	projectConfig := &CreateNewProjectConfig{
		CreateProjectBody: management_api.CreateProjectBody{
			Name:       c.ProjectName,
			Plan:       "free",
			KpsEnabled: false,
		},
	}

	if err := promptOrganizations(projectConfig); err != nil {
		return nil, err
	}

	if err := promptRegion(projectConfig); err != nil {
		return nil, err
	}

	if err := promptDbPassword(projectConfig); err != nil {
		return nil, err
	}

	return projectConfig, nil
}

func promptOrganizations(cc *CreateNewProjectConfig) error {
	// todo : create new organization if doesn`t have organization
	orgs, err := GetOrganizationOptions()
	if err != nil {
		return err
	}
	input := selection.New("Select organization", orgs.ToFlatName())
	input.PageSize = 4

	inputText, err := input.RunPrompt()
	if err != nil {
		return err
	}

	org, isFound := orgs.FindByName(inputText)
	if !isFound {
		return errors.New("no available organization")
	}

	cc.OrganizationId = org.Id
	return nil
}

func promptRegion(cc *CreateNewProjectConfig) error {
	input := selection.New("Select regions", supabase.AllRegion)
	input.PageSize = 4

	inputText, err := input.RunPrompt()
	if err != nil {
		return err
	}

	cc.Region = string(inputText)
	return nil
}

func promptDbPassword(cc *CreateNewProjectConfig) error {
	input := textinput.New("Set database password")
	input.Validate = func(s string) error {
		if s == "" {
			return errors.New("database password can`t be empty")
		}
		return nil
	}
	input.Hidden = true
	input.HideMask = '*'

	inputText, err := input.RunPrompt()
	if err != nil {
		return err
	}

	cc.DbPass = inputText
	return nil
}

// ----- Check and generate functionality -----

func Generate(config *raiden.Config, projectPath string) error {
	return generator.GenerateConfig(projectPath, config, generator.Generate)
}

func GetConfigFilePath(projectPath string) string {
	return filepath.Join(projectPath, generator.ConfigDir, fmt.Sprintf("%s.yaml", generator.ConfigFile))
}

func IsConfigExist(projectPath string) bool {
	return utils.IsFileExists(GetConfigFilePath(projectPath))
}
