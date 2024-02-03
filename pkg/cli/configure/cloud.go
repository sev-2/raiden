package configure

import (
	"fmt"

	"github.com/sev-2/raiden/pkg/supabase"
)

// This file contain functionality to interact with supabase cloud

// ----- Bind Project Id And Public For Cloud Deployment -----
func BindProjectIpAndPublicUrl(c *Config) (isExist bool, err error) {
	supabase.ConfigureManagementApi(c.SupabaseApiUrl, c.AccessToken)
	project, err := supabase.FindProject(c.ProjectName)
	if err != nil {
		return false, err
	}

	if project.Id == "" {
		return false, err
	}

	c.ProjectId = project.Id
	c.SupabasePublicUrl = fmt.Sprintf("https://%s.supabase.co", project.Id)

	return true, nil
}

// ----- Supabase Organizations -----
type Organizations struct {
	supabase.Organizations
}

func (o *Organizations) ToFlatName() []string {
	var organizationOptions []string
	if o.Organizations == nil {
		return organizationOptions
	}

	for _, org := range o.Organizations {
		organizationOptions = append(organizationOptions, org.Name)
	}

	return organizationOptions
}

func (o *Organizations) FindByName(name string) (org *supabase.Organization, isFound bool) {
	if o.Organizations == nil {
		return nil, false
	}

	for _, org := range o.Organizations {
		if org.Name == name {
			return (*supabase.Organization)(&org), true
		}
	}

	return nil, false
}

func GetOrganizationOptions() (*Organizations, error) {
	orgs, err := supabase.GetOrganizations()
	if err != nil {
		return nil, err
	}

	return &Organizations{orgs}, nil
}

// ----- Supabase Project -----
func CreateNewProject(cc *CreateNewProjectConfig) (supabase.Project, error) {
	return supabase.CreateNewProject(cc.CreateProjectBody)
}
