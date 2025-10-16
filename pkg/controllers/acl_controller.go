package controllers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/acl"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/valyala/fasthttp"
)

// User role controller

type AclUserRoleRequest struct {
	UserUuid string `json:"user_uuid" validate:"requiredForMethod=Patch"`
	Role     string `json:"role" validate:"requiredForMethod=Patch"`
}

type AclUserRoleResponse struct {
	Message string `json:"message"`
}

type AclUserRoleController struct {
	raiden.ControllerBase
	Http    string `path:"/acl/user/role" type:"custom"`
	Payload *AclUserRoleRequest
	Result  AclUserRoleResponse
}

func (c *AclUserRoleController) Patch(ctx raiden.Context) error {
	if _, errAccess := acl.GetAuthenticatedData(ctx, true); errAccess != nil {
		return errAccess
	}

	role, err := acl.GetRole(c.Payload.Role)
	if err != nil {
		raiden.Error("failed to get role", "message", err)
		return ctx.SendError(err.Error())
	}

	if err = acl.SetUserRole(ctx.Config(), c.Payload.UserUuid, role); err != nil {
		raiden.Error("failed to set role", "message", err)
		return ctx.SendError("failed to User role")
	}
	c.Result.Message = fmt.Sprintf("success update role user to %s", c.Payload.Role)
	return ctx.SendJson(c.Result)
}

type AclRoleRequest struct{}

type AclRoleResponse []AclRoleItemResponse

type AclRoleItemResponse struct {
	Name         string                `json:"name"`
	InheritRoles []AclRoleItemResponse `json:"inherit_roles"`
}

type AclRoleController struct {
	raiden.ControllerBase
	Http    string `path:"/acl/roles" type:"custom"`
	Payload *AclRoleRequest
	Result  AclRoleResponse
}

func (c *AclRoleController) Get(ctx raiden.Context) error {
	if _, errAccess := acl.GetAuthenticatedData(ctx, true); errAccess != nil {
		return errAccess
	}

	roles, err := acl.GetRoles()
	if err != nil {
		raiden.Error("failed to get roles", "message", err)
		return ctx.SendError("failed to get roles")
	}

	for _, r := range roles {
		inheritRoles := make([]AclRoleItemResponse, 0)
		if len(r.InheritRoles()) > 0 {
			for _, ir := range r.InheritRoles() {
				inheritRoles = append(inheritRoles, AclRoleItemResponse{
					Name: ir.Name(),
				})
			}
		}

		c.Result = append(c.Result, AclRoleItemResponse{
			Name:         r.Name(),
			InheritRoles: inheritRoles,
		})
	}

	return ctx.SendJson(c.Result)
}

type AclUserPolicyRequest struct{}

type AclUserPolicyResponse []string

type AclUserPolicyController struct {
	raiden.ControllerBase
	Http    string `path:"/acl/user/policies" type:"custom"`
	Payload *AclUserPolicyRequest
	Result  AclUserPolicyResponse
}

func (c *AclUserPolicyController) Get(ctx raiden.Context) error {
	data, errAccess := acl.GetAuthenticatedData(ctx, false)
	if errAccess != nil {
		return errAccess
	}

	roles, err := acl.GetRoles()
	if err != nil {
		return ctx.SendError("failed to validate role")
	}

	mapRole := make(map[string]raiden.Role)
	for _, rr := range roles {
		mapRole[rr.Name()] = rr
	}

	filterRole := make([]string, 0)
	if data.Role != acl.ServiceRoleName {
		fr, exist := mapRole[data.Role]
		if !exist {
			return ctx.SendErrorWithCode(fasthttp.StatusForbidden, errors.New("the token is no longer valid"))
		}
		filterRole = append(filterRole, fr.Name())
		ih := fr.InheritRoles()
		if len(ih) > 0 {
			for _, ihi := range ih {
				filterRole = append(filterRole, ihi.Name())
			}
		}
	}

	policies, err := acl.GetPolicy(filterRole)
	if err != nil {
		raiden.Error("failed to get permission", "message", err)
		return ctx.SendError("failed to get policies")
	}

	for _, p := range policies {
		key := fmt.Sprintf("%s.%s.%s", strings.ToLower(p.Table), strings.ToLower(string(p.Command)), utils.ToSnakeCase(p.Name))
		c.Result = append(c.Result, key)
	}

	if c.Result == nil {
		c.Result = make(AclUserPolicyResponse, 0)
	}

	return ctx.SendJson(c.Result)
}
