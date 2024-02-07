package state

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/importer"
	"go/token"
	"go/types"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/utils"
)

func ToRoles(roleStates []RoleState, withNativeRole bool) (roles []supabase.Role, err error) {
	var paths []string
	for i := range roleStates {
		r := roleStates[i]
		if !r.IsNative {
			paths = append(paths, r.RolePath)
		}
	}

	fset, astFiles, err := loadFiles(paths)
	if err != nil {
		return roles, err
	}

	conf := types.Config{}
	conf.Importer = importer.Default()
	pkg, err := conf.Check("roles", fset, astFiles, nil)
	if err != nil {
		fmt.Println("Error type-checking code:", err)
		return
	}

	for i := range roleStates {
		r := roleStates[i]

		if r.IsNative && withNativeRole {
			roles = append(roles, r.Role)
		}

		if !r.IsNative {
			sr, err := createRoleFromState(pkg, astFiles, fset, r)
			if err != nil {
				return roles, err
			}
			roles = append(roles, sr)
		}
	}

	return
}

func createRoleFromState(pkg *types.Package, astFiles []*ast.File, fset *token.FileSet, state RoleState) (role supabase.Role, err error) {
	obj := pkg.Scope().Lookup(state.RoleStruct)
	if obj == nil {
		fmt.Println("Struct not found:", state.RoleStruct)
		return
	}

	// Assert the object's type to *types.TypeName
	typeObj, ok := obj.(*types.TypeName)
	if !ok {
		fmt.Println("Unexpected type for object:", obj)
		return
	}

	// Get the reflect.Type of the struct
	structType := typeObj.Type().Underlying().(*types.Struct)
	role = state.Role
	role.Name = utils.ToSnakeCase(typeObj.Name())

	// Iterate over the fields of the struct
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		fieldName := field.Name()
		fieldTag := structType.Tag(i)

		if fieldName == "Metadata" {
			metaTag := raiden.UnmarshalRoleMetadataTag(fieldTag)

			if metaTag.Name != "" {
				role.Name = metaTag.Name
			}

			if metaTag.Config != nil {
				role.Config = metaTag.Config
			}

			role.ConnectionLimit = metaTag.ConnectionLimit
			role.InheritRole = metaTag.InheritRole
			role.IsReplicationRole = metaTag.IsReplicationRole
			role.IsSuperuser = metaTag.IsSuperuser
		}

		if fieldName == "Permission" {
			permissionTag := raiden.UnmarshalRolePermissionTag(fieldTag)
			role.CanBypassRLS = permissionTag.CanBypassRls
			role.CanCreateDB = permissionTag.CanCreateDB
			role.CanCreateRole = permissionTag.CanCreateRole
			role.CanLogin = permissionTag.CanLogin
		}
	}

	return role, nil
}

func CompareRoles(supabaseRoles []supabase.Role, appRoles []supabase.Role) (diffResult []CompareDiffResult, err error) {
	mapAppRoles := make(map[int]supabase.Role)
	for i := range appRoles {
		r := appRoles[i]
		mapAppRoles[r.ID] = r
	}

	for i := range supabaseRoles {
		r := supabaseRoles[i]

		appRole, isExist := mapAppRoles[r.ID]
		if isExist {
			spByte, err := json.Marshal(r)
			if err != nil {
				return diffResult, err
			}
			spHash := utils.HashByte(spByte)

			appByte, err := json.Marshal(appRole)
			if err != nil {
				return diffResult, err
			}
			appHash := utils.HashByte(appByte)

			if spHash != appHash {
				diffResult = append(diffResult, CompareDiffResult{
					Name:             r.Name,
					Category:         CompareDiffCategoryConflict,
					SupabaseResource: r,
					AppResource:      appRole,
				})
			}
		}
	}

	return
}
