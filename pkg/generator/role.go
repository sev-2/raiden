package generator

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

var RoleLogger hclog.Logger = logger.HcLog().Named("generator.role")

// ----- Define type, variable and constant -----
type GenerateRoleData struct {
	Imports                []string
	Package                string
	Name                   string
	DefaultLimitConnection int
	ConnectionLimit        int
	InheritRole            bool
	IsReplicationRole      bool
	IsSuperuser            bool
	CanBypassRls           bool
	CanCreateDB            bool
	CanCreateRole          bool
	CanLogin               bool
	ValidUntil             string
	InheritRoles           string
}

const (
	RoleDir      = "internal/roles"
	RoleTemplate = `package {{ .Package }}
{{- if gt (len .Imports) 0 }}

import (
{{- range .Imports}}
	{{.}}
{{- end}}
)
{{- end }}

type {{ .Name | ToGoIdentifier }} struct {
	raiden.RoleBase
}

func (r *{{ .Name | ToGoIdentifier }}) Name() string {
	return "{{ .Name }}"
}

{{- if ne .ConnectionLimit .DefaultLimitConnection }}

func (r *{{ .Name | ToGoIdentifier }}) ConnectionLimit() int {
	return {{ .ConnectionLimit }}
}
{{- end }}
{{- if not .InheritRole }}

func (r *{{ .Name | ToGoIdentifier }}) IsInheritRole() bool {
	return {{ .InheritRole }}
}
{{- end }}
{{- if ne .InheritRoles "" }}

func (r *{{ .Name | ToGoIdentifier }}) InheritRoles() []raiden.Role {
	return []raiden.Role{ {{ .InheritRoles }} }
}
{{- end }}
{{- if .IsReplicationRole }}

func (r *{{ .Name | ToGoIdentifier }}) IsReplicationRole() bool {
	return {{ .IsReplicationRole }}
}
{{- end }}
{{- if .IsSuperuser }}
func (r *{{ .Name | ToGoIdentifier }}) IsSuperuser() bool {
	return {{ .IsSuperuser }}
}

{{- end }}
{{- if .CanBypassRls }}

func (r *{{ .Name | ToGoIdentifier }}) CanBypassRls() bool {
	return {{ .CanBypassRls }}
}
{{- end }}
{{- if .CanCreateDB }}

func (r *{{ .Name | ToGoIdentifier }}) CanCreateDB() bool {
	return {{ .CanCreateDB }}
}
{{- end }}
{{- if .CanCreateRole }}

func (r *{{ .Name | ToGoIdentifier }}) CanCreateRole() bool {
	return {{ .CanCreateRole }}
}
{{- end }}
{{- if .CanLogin }}

func (r *{{ .Name | ToGoIdentifier }}) CanLogin() bool {
	return {{ .CanLogin }}
}
{{- end }}
{{- if ne .ValidUntil ""}}

func (r *{{ .Name | ToGoIdentifier }}) ValidUntil() *objects.SupabaseTime {
	t, err := time.Parse(raiden.DefaultRoleValidUntilLayout, "{{ .ValidUntil }}")
	if err != nil {
		raiden.Error(err.Error())
		return nil
	}
	return objects.NewSupabaseTime(t)
}
{{- end }}
`
)

func GenerateRoles(basePath string, roles []objects.Role, generateFn GenerateFn) (err error) {
	folderPath := filepath.Join(basePath, RoleDir)
	RoleLogger.Trace("create roles folder if not exist", folderPath)
	if exist := utils.IsFolderExists(folderPath); !exist {
		if err := utils.CreateFolder(folderPath); err != nil {
			return err
		}
	}

	for _, v := range roles {
		if err := GenerateRole(folderPath, v, generateFn); err != nil {
			return err
		}
	}

	return nil
}

func GenerateRole(folderPath string, role objects.Role, generateFn GenerateFn) error {
	// define binding func
	funcMaps := []template.FuncMap{
		{"ToGoIdentifier": utils.SnakeCaseToPascalCase},
	}

	// define file path
	filePath := filepath.Join(folderPath, fmt.Sprintf("%s.%s", role.Name, "go"))

	// set imports path
	var imports []string
	raidenPath := fmt.Sprintf("%q", "github.com/sev-2/raiden")
	imports = append(imports, raidenPath)

	var validUntil string
	if role.ValidUntil != nil {
		imports = append(
			imports,
			fmt.Sprintf("%q", "time"),
			fmt.Sprintf("%q", "github.com/sev-2/raiden/pkg/supabase/objects"),
		)
		validUntil = role.ValidUntil.Format(raiden.DefaultRoleValidUntilLayout)
	}

	// set inherit roles
	var inheritRoles []string
	if role.InheritRole && len(role.InheritRoles) > 0 {
		for _, r := range role.InheritRoles {
			inheritRoles = append(inheritRoles, fmt.Sprintf("&%s{}", utils.SnakeCaseToPascalCase(r.Name)))
		}
	}

	// execute the template and write to the file
	data := GenerateRoleData{
		Package:                "roles",
		Imports:                imports,
		Name:                   role.Name,
		ConnectionLimit:        role.ConnectionLimit,
		DefaultLimitConnection: raiden.DefaultRoleConnectionLimit,
		InheritRole:            role.InheritRole,
		InheritRoles:           strings.Join(inheritRoles, ","),
		IsReplicationRole:      role.IsReplicationRole,
		IsSuperuser:            role.IsSuperuser,
		CanBypassRls:           role.CanBypassRLS,
		CanCreateDB:            role.CanCreateDB,
		CanCreateRole:          role.CanCreateRole,
		CanLogin:               role.CanLogin,
		ValidUntil:             validUntil,
	}

	// set input
	input := GenerateInput{
		BindData:     data,
		Template:     RoleTemplate,
		TemplateName: "roleTemplate",
		OutputPath:   filePath,
		FuncMap:      funcMaps,
	}

	RoleLogger.Debug("generate role", "path", input.OutputPath)
	return generateFn(input, nil)
}
