/*
 * Supabase API (v1)
 *
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * API version: 1.0.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package management_api

type V1OrganizationSlugResponse struct {
	Plan      *BillingPlanId `json:"plan,omitempty"`
	OptInTags []string       `json:"opt_in_tags"`
	Id        string         `json:"id"`
	Name      string         `json:"name"`
}