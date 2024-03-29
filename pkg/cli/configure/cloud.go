package configure

import (
	"fmt"

	"github.com/sev-2/raiden/pkg/supabase"
)

// This file contain functionality to interact with supabase cloud

// ----- Bind Project Id And Public For Cloud Deployment -----
func BindProjectIpAndPublicUrl(c *Config) (isExist bool, err error) {
	project, err := supabase.FindProject(&c.Config)
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
