package raiden

import (
	"strings"
	"sync"

	"github.com/sev-2/raiden/pkg/builder"
	"github.com/sev-2/raiden/pkg/supabase/constants"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

type (
	Acl struct {
		// New Handle Acl
		Schema   string
		Table    string
		isEnable bool
		isForced bool

		mapRule map[string]*rule

		once sync.Once
	}
)

type Command string

const (
	CommandAll    Command = "ALL"
	CommandSelect Command = "SELECT"
	CommandInsert Command = "INSERT"
	CommandUpdate Command = "UPDATE"
	CommandDelete Command = "DELETE"
)

func (m Command) ToSupabaseCommand() objects.PolicyCommand {
	switch m {
	case CommandSelect:
		return objects.PolicyCommandSelect
	case CommandInsert:
		return objects.PolicyCommandInsert
	case CommandUpdate:
		return objects.PolicyCommandUpdate
	case CommandDelete:
		return objects.PolicyCommandDelete
	case CommandAll:
		return objects.PolicyCommandAll
	}

	return ""
}

type AclMode int

const (
	AclModePermissive AclMode = iota // default
	AclModeRestrictive
)

func (m AclMode) ActionString() string {
	if m == AclModeRestrictive {
		return "RESTRICTIVE"
	}
	return "PERMISSIVE"
}

func (a *Acl) Enable() *Acl {
	a.isEnable = true
	return a
}

func (a *Acl) Forced() *Acl {
	a.isForced = true
	return a
}

func (a *Acl) IsEnable() bool {
	return a.isEnable
}

func (a *Acl) IsForced() bool {
	return a.isForced
}

func (a *Acl) Define(rr ...*rule) *Acl {
	if a.mapRule == nil {
		a.mapRule = make(map[string]*rule, 0)
	}

	for _, r := range rr {
		a.mapRule[r.name] = r
	}
	return a
}

func (a *Acl) InitOnce(fn func()) { a.once.Do(fn) }

func (a *Acl) BuildPolicies(schema, table string) (policies objects.Policies, err error) {
	a.Schema = schema
	a.Table = table

	// build rule
	for _, r := range a.mapRule {
		if p, err := r.Build(a.Schema, a.Table); err != nil {
			return policies, err
		} else {
			policies = append(policies, *p)
		}
	}

	return policies, nil
}

func (a *Acl) BuildStoragePolicies(bucketName string) (objects.Policies, error) {
	bucketName = strings.TrimSpace(bucketName)
	if err := utils.EmptyOrError(bucketName, "bucket name required to set"); err != nil {
		return nil, err
	}

	a.Schema = constants.DefaultStorageSchema
	a.Table = constants.DefaultObjectTable

	policies := make(objects.Policies, 0, len(a.mapRule))
	for _, r := range a.mapRule {
		policy, err := r.buildWithClauses(a.Schema, a.Table,
			builder.StorageUsingClause(bucketName, r.using),
			builder.StorageCheckClause(bucketName, r.check),
		)
		if err != nil {
			return nil, err
		}
		policies = append(policies, *policy)
	}

	return policies, nil
}

func Rule(name string) *rule {
	return &rule{
		name: name,
	}
}

type rule struct {
	name    string
	command Command
	roles   []string
	using   builder.Clause
	check   builder.Clause
	mode    AclMode
}

func (r *rule) For(roles ...string) *rule {
	r.roles = append(r.roles, roles...)
	return r
}

func (r *rule) To(command Command) *rule {
	r.command = command
	return r
}

func (r *rule) Using(clause builder.Clause) *rule {
	r.using = clause
	return r
}

func (r *rule) Check(clause builder.Clause) *rule {
	r.check = clause
	return r
}

func (r *rule) WithPermissive() *rule {
	r.mode = AclModePermissive
	return r
}

func (r *rule) WithRestrictive() *rule {
	r.mode = AclModeRestrictive
	return r
}

func (r *rule) Build(schema string, table string) (*objects.Policy, error) {
	return r.buildWithClauses(schema, table, r.using, r.check)
}

func (r *rule) buildWithClauses(schema string, table string, usingClause, checkClause builder.Clause) (*objects.Policy, error) {
	// Validate Required Value
	if err := utils.EmptyOrError(r.name, "rule name required to set"); err != nil {
		return nil, err
	}

	if err := utils.EmptyOrError(table, "table name required to set"); err != nil {
		return nil, err
	}

	// Set default value
	schema = utils.EmptyOrDefault(schema, "public")
	mode := utils.EmptyOrDefault(r.mode, AclModePermissive)

	usingRaw := strings.TrimSpace(usingClause.String())
	checkRaw := strings.TrimSpace(checkClause.String())

	definition, checkPtr := r.prepareClauses(usingRaw, checkRaw)

	return &objects.Policy{
		Schema:     schema,
		Table:      table,
		Name:       r.name,
		Action:     mode.ActionString(), // "PERMISSIVE"/"RESTRICTIVE"
		Roles:      r.roles,
		Command:    r.command.ToSupabaseCommand(),
		Definition: definition,
		Check:      checkPtr,
	}, nil
}

func (r *rule) prepareClauses(usingRaw, checkRaw string) (definition string, checkPtr *string) {
	switch r.command {
	case CommandSelect:
		definition = ensureClause(usingRaw, true)
		return definition, nil
	case CommandInsert:
		definition = ""
		return definition, ensureCheck(checkRaw, true)
	case CommandDelete:
		merged := strings.TrimSpace(usingRaw)
		if merged == "" {
			merged = checkRaw
		}
		definition = ensureClause(merged, true)
		return definition, nil
	case CommandUpdate:
		definition = ensureClause(usingRaw, true)
		return definition, ensureCheck(checkRaw, true)
	case CommandAll:
		definition = ensureClause(usingRaw, true)
		return definition, ensureCheck(checkRaw, true)
	default:
		definition = ensureClause(usingRaw, true)
		return definition, ensureCheck(checkRaw, false)
	}
}

func ensureClause(value string, requireFallback bool) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" && requireFallback {
		return "TRUE"
	}
	return trimmed
}

func ensureCheck(value string, defaultTrue bool) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		if !defaultTrue {
			return nil
		}
		trimmed = "TRUE"
	}
	return &trimmed
}
