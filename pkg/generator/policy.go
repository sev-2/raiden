package generator

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

type Rls struct {
	CanWrite []string
	CanRead  []string
}

func BuildRlsTag(rlsList objects.Policies, name string, rlsType supabase.RlsType) string {
	var rls Rls

	var readUsingTag, writeCheckTag, writeUsingTag string
	for _, v := range rlsList {
		switch v.Command {
		case objects.PolicyCommandSelect:
			if v.Name == supabase.GetPolicyName(objects.PolicyCommandSelect, strings.ToLower(string(rlsType)), name) {
				rls.CanRead = append(rls.CanRead, v.Roles...)
				if v.Definition != "" {
					readUsingTag = v.Definition
				}
			}
		case objects.PolicyCommandInsert, objects.PolicyCommandUpdate, objects.PolicyCommandDelete:
			if v.Name == supabase.GetPolicyName(objects.PolicyCommandInsert, strings.ToLower(string(rlsType)), name) {
				if len(rls.CanWrite) == 0 {
					rls.CanWrite = append(rls.CanWrite, v.Roles...)
				}

				if len(writeCheckTag) == 0 && v.Check != nil {
					writeCheckTag = *v.Check
				}
			}

			if v.Name == supabase.GetPolicyName(objects.PolicyCommandUpdate, strings.ToLower(string(rlsType)), name) && len(rls.CanWrite) == 0 {
				if len(rls.CanWrite) == 0 {
					rls.CanWrite = append(rls.CanWrite, v.Roles...)
				}

				if len(writeCheckTag) == 0 && v.Check != nil {
					writeCheckTag = *v.Check
				}

				if len(writeUsingTag) == 0 && v.Definition != "" {
					writeUsingTag = v.Definition
				}
			}

			if v.Name == supabase.GetPolicyName(objects.PolicyCommandDelete, strings.ToLower(string(rlsType)), name) && len(rls.CanWrite) == 0 {
				if len(rls.CanWrite) == 0 {
					rls.CanWrite = append(rls.CanWrite, v.Roles...)
				}

				if len(writeUsingTag) == 0 && v.Definition != "" {
					writeUsingTag = v.Definition
				}
			}
		}
	}

	rlsTag := fmt.Sprintf("read:%q write:%q", strings.Join(rls.CanRead, ","), strings.Join(rls.CanWrite, ","))
	if len(readUsingTag) > 0 {
		cleanTag := strings.TrimLeft(strings.TrimRight(readUsingTag, ")"), "(")
		if rlsType == supabase.RlsTypeStorage {
			cleanTag = cleanupRlsTagStorage(name, cleanTag)
		}

		if cleanTag != "" {
			rlsTag = fmt.Sprintf("%s readUsing:%q", rlsTag, cleanTag)
		}
	}

	if len(writeCheckTag) > 0 {
		cleanTag := strings.TrimLeft(strings.TrimRight(writeCheckTag, ")"), "(")
		if rlsType == supabase.RlsTypeStorage {
			cleanTag = cleanupRlsTagStorage(name, cleanTag)
		}

		if cleanTag != "" {
			rlsTag = fmt.Sprintf("%s writeCheck:%q", rlsTag, cleanTag)
		}
	}

	if len(writeUsingTag) > 0 {
		cleanTag := strings.TrimLeft(strings.TrimRight(writeUsingTag, ")"), "(")
		if rlsType == supabase.RlsTypeStorage {
			cleanTag = cleanupRlsTagStorage(name, cleanTag)
		}

		if cleanTag != "" {
			rlsTag = fmt.Sprintf("%s writeUsing:%q", rlsTag, cleanTag)
		}
	}

	return rlsTag
}

func cleanupRlsTagStorage(name, tag string) string {
	// clean storage identifier
	cleanTag := strings.Replace(tag, fmt.Sprintf("bucket_id = '%s'", name), "", 1)
	cleanTag = strings.Replace(cleanTag, "AND", "", 1)
	cleanTag = strings.Replace(cleanTag, "OR", "", 1)
	cleanTag = strings.TrimLeftFunc(cleanTag, unicode.IsSpace)
	return cleanTag
}

var PolicyLogger hclog.Logger = logger.HcLog().Named("generator.policy")
