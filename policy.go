package raiden

import (
	"errors"
	"regexp"
	"strings"

	"github.com/sev-2/raiden/pkg/builder"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

type (
	Acl struct {
		Roles []string
		Check *string
		Using string
	}

	AclTag struct {
		Read  Acl
		Write Acl
	}
)

func UnmarshalAclTag(tag string) AclTag {
	var aclTag AclTag

	aclTagMap := make(map[string]string)

	// Regular expression to match key-value pairs
	re := regexp.MustCompile(`(\w+):"([^"]*)"`)

	// Find all matches
	matches := re.FindAllStringSubmatch(tag, -1)

	// Loop through matches and add to result map
	for _, match := range matches {
		if len(match) == 3 {
			key := match[1]
			value := match[2]
			aclTagMap[key] = value
		}
	}

	if readTag, exist := aclTagMap["read"]; exist && len(readTag) > 0 {
		aclTag.Read.Roles = strings.Split(readTag, ",")
	}

	if writeTag, exist := aclTagMap["write"]; exist && len(writeTag) > 0 {
		aclTag.Write.Roles = strings.Split(writeTag, ",")
	}

	if readTagUsing, exist := aclTagMap["readUsing"]; exist && len(readTagUsing) > 0 {
		aclTag.Read.Using = readTagUsing
	}

	if writeTagCheck, exist := aclTagMap["writeCheck"]; exist && len(writeTagCheck) > 0 {
		aclTag.Write.Check = &writeTagCheck
	}

	if writeTagUsing, exist := aclTagMap["writeUsing"]; exist && len(writeTagUsing) > 0 {
		aclTag.Write.Using = writeTagUsing
	}

	return aclTag
}

// -----------------------------------------
// Policy Resources
// -----------------------------------------

type PolicyCommand string

const (
	PolicyCommandAll    PolicyCommand = "ALL"
	PolicyCommandSelect PolicyCommand = "SELECT"
	PolicyCommandInsert PolicyCommand = "INSERT"
	PolicyCommandUpdate PolicyCommand = "UPDATE"
	PolicyCommandDelete PolicyCommand = "DELETE"
)

func (m PolicyCommand) ToSupabaseCommand() objects.PolicyCommand {
	switch m {
	case PolicyCommandSelect:
		return objects.PolicyCommandSelect
	case PolicyCommandInsert:
		return objects.PolicyCommandInsert
	case PolicyCommandUpdate:
		return objects.PolicyCommandUpdate
	case PolicyCommandDelete:
		return objects.PolicyCommandDelete
	case PolicyCommandAll:
		return objects.PolicyCommandAll
	}

	return ""
}

type PolicyMode int

const (
	ModePermissive PolicyMode = iota // default
	ModeRestrictive
)

func (m PolicyMode) ActionString() string {
	if m == ModeRestrictive {
		return "RESTRICTIVE"
	}
	return "PERMISSIVE"
}

type PolicyResourceType string

const (
	PolicyResourceTypeModel   PolicyCommand = "TABLE"
	PolicyResourceTypeStorage PolicyCommand = "STORAGE"
)

type (
	Policy interface {
		Name() string
		Model() any
		Storage() Bucket
		Roles() []Role
		Command() PolicyCommand
		Mode() PolicyMode
		Definition() builder.Clause
		Check() builder.Clause
	}

	PolicyBase struct {
	}
)

var (
	DefaultPolicySchema = "public"
)

func (p PolicyBase) Name() string {
	return ""
}

func (p PolicyBase) Model() any {
	return nil
}
func (p PolicyBase) Storage() Bucket {
	return nil
}
func (p PolicyBase) Roles() []Role {
	return []Role{}
}
func (p PolicyBase) Command() PolicyCommand {
	return ""
}
func (p PolicyBase) Mode() PolicyMode {
	return ModePermissive
}
func (p PolicyBase) Definition() builder.Clause {
	return ""
}
func (p PolicyBase) Check() builder.Clause {
	return ""
}

func BuildPolicy(p Policy) (*objects.Policy, error) {
	// Validate Required Value
	if err := utils.EmptyOrError(p.Name(), "name required to set"); err != nil {
		return nil, err
	}

	if p.Model() == nil && p.Storage() == nil {
		return nil, errors.New("policy must define resouce, implement Model() or Storage() function")
	}

	if err := utils.EmptyOrError(p.Command(), "command is rquired to set"); err != nil {
		return nil, err
	}

	// TABLE and SCHEMA
	var schema, table string

	if p.Model() != nil {
		schema, table = builder.TableFromModel(p.Model())
		schema = utils.EmptyOrDefault(schema, DefaultPolicySchema)
		if err := utils.EmptyOrError(table, "invalid model, table is not recognize"); err != nil {
			return nil, err
		}
	}

	if p.Storage() != nil {
		schema, table = "storage", p.Storage().Name()
		if err := utils.EmptyOrError(table, "invalid storage, bucket is not recognize"); err != nil {
			return nil, err
		}
	}

	// USING (definition)
	def := strings.TrimSpace(p.Definition().String())

	// WITH CHECK (optional)
	var checkPtr *string
	if c := strings.TrimSpace(p.Check().String()); c != "" {
		checkPtr = &c
	}

	// MODE
	mode := utils.EmptyOrDefault(p.Mode(), ModePermissive)

	// Roles
	roles := utils.EmptyOrDefault(p.Roles(), []Role{})
	rolesArr := []string{}
	if len(roles) > 0 {
		for _, r := range roles {
			if r != nil {
				rolesArr = append(rolesArr, r.Name())
			}
		}
	}

	return &objects.Policy{
		Schema:     schema,
		Table:      table,
		Name:       p.Name(),
		Action:     mode.ActionString(), // "PERMISSIVE"/"RESTRICTIVE"
		Roles:      rolesArr,
		Command:    p.Command().ToSupabaseCommand(),
		Definition: def,
		Check:      checkPtr,
	}, nil
}
