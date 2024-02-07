package raiden

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/utils"
)

type (
	RoleMetadataTag struct {
		Name              string
		Config            map[string]any
		ConnectionLimit   int
		InheritRole       bool
		IsReplicationRole bool
		IsSuperuser       bool
	}

	RolePermissionTag struct {
		CanBypassRls  bool
		CanCreateDB   bool
		CanCreateRole bool
		CanLogin      bool
	}

	Role supabase.Role
)

const (
	DefaultRoleConnectionLimit = 60
)

func UnmarshalRole(role any) (r Role, err error) {
	roleType := reflect.TypeOf(role)

	if roleType.Kind() == reflect.Ptr || roleType.Kind() == reflect.Pointer {
		roleType = roleType.Elem()
	}

	metadataField, isFound := roleType.FieldByName("Metadata")
	if !isFound {
		return r, fmt.Errorf("field Metadata is not exist in %s", roleType.Name())
	}

	permissionField, isFound := roleType.FieldByName("Permission")
	if !isFound {
		return r, fmt.Errorf("field Permission is not exist in %s", roleType.Name())
	}

	metadataTag := UnmarshalRoleMetadataTag(fmt.Sprintf("%v", metadataField.Tag))
	permissionTag := UnmarshalRolePermissionTag(fmt.Sprintf("%v", permissionField.Tag))

	if metadataTag.Name == "" {
		metadataTag.Name = utils.ToSnakeCase(roleType.Name())
	}

	r = Role{
		CanBypassRLS:      permissionTag.CanBypassRls,
		CanCreateDB:       permissionTag.CanCreateDB,
		CanCreateRole:     permissionTag.CanCreateRole,
		CanLogin:          permissionTag.CanLogin,
		Config:            metadataTag.Config,
		ConnectionLimit:   metadataTag.ConnectionLimit,
		InheritRole:       metadataTag.InheritRole,
		IsReplicationRole: metadataTag.IsReplicationRole,
		IsSuperuser:       metadataTag.IsSuperuser,
		Name:              metadataTag.Name,
	}

	return r, nil
}

func UnmarshalRoleMetadataTag(rawTag string) RoleMetadataTag {
	var metadata RoleMetadataTag
	tagMap := utils.ParseTag(rawTag)
	if configTag, exist := tagMap["config"]; exist {
		configSplit := strings.Split(configTag, ";")
		configMap := make(map[string]any)
		for _, c := range configSplit {
			cSplit := strings.Split(c, ":")
			if len(cSplit) == 2 {
				configMap[cSplit[0]] = cSplit[1]
			}
		}
		metadata.Config = configMap
	}

	if nameTag, exist := tagMap["name"]; exist {
		metadata.Name = nameTag
	}

	metadata.ConnectionLimit, _ = strconv.Atoi(tagMap["connectionLimit"])
	metadata.InheritRole = utils.ParseBool(tagMap["inheritRole"])
	metadata.IsReplicationRole = utils.ParseBool(tagMap["isReplicationRole"])
	metadata.IsSuperuser = utils.ParseBool(tagMap["isSuperuser"])
	return metadata
}

func MarshalRoleMetadataTag(metadataTag *RoleMetadataTag) string {
	if metadataTag == nil {
		return ""
	}

	var tagArr []string

	// append config
	if metadataTag.Name != "" {
		tagArr = append(tagArr, fmt.Sprintf("name:%q", metadataTag.Name))
	}

	if metadataTag.Config != nil {
		var configTagArr []string
		for k, v := range metadataTag.Config {
			configTagArr = append(configTagArr, fmt.Sprintf("%s:%s", k, v))
		}
		configTag := strings.Join(configTagArr, ";")
		tagArr = append(tagArr, fmt.Sprintf("config:%q", configTag))
	}

	if metadataTag.ConnectionLimit != 0 {
		tagArr = append(tagArr, fmt.Sprintf("connectionLimit:%q", strconv.Itoa(metadataTag.ConnectionLimit)))
	} else {
		tagArr = append(tagArr, fmt.Sprintf("connectionLimit:%q", DefaultRoleConnectionLimit))
	}

	tagArr = append(tagArr,
		fmt.Sprintf("inheritRole:%q", strconv.FormatBool(metadataTag.InheritRole)),
		fmt.Sprintf("isReplicationRole:%q", strconv.FormatBool(metadataTag.IsReplicationRole)),
		fmt.Sprintf("isSuperuser:%q", strconv.FormatBool(metadataTag.IsSuperuser)),
	)

	return strings.Join(tagArr, " ")
}

func UnmarshalRolePermissionTag(rawTag string) RolePermissionTag {
	var permission RolePermissionTag
	tagMap := utils.ParseTag(rawTag)
	permission.CanBypassRls = utils.ParseBool(tagMap["canBypassRls"])
	permission.CanCreateDB = utils.ParseBool(tagMap["canCreateDb"])
	permission.CanCreateRole = utils.ParseBool(tagMap["canCreateRole"])
	permission.CanLogin = utils.ParseBool(tagMap["canLogin"])
	return permission
}

func MarshalRolePermissionTag(permissionTag *RolePermissionTag) string {
	if permissionTag == nil {
		return ""
	}
	var tagArr []string
	tagArr = append(tagArr,
		fmt.Sprintf("canBypassRls:%q", strconv.FormatBool(permissionTag.CanBypassRls)),
		fmt.Sprintf("canCreateDb:%q", strconv.FormatBool(permissionTag.CanCreateDB)),
		fmt.Sprintf("canCreateRole:%q", strconv.FormatBool(permissionTag.CanCreateRole)),
		fmt.Sprintf("canLogin:%q", strconv.FormatBool(permissionTag.CanLogin)),
	)
	return strings.Join(tagArr, " ")
}
