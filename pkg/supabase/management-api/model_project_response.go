/*
 * Supabase API (v1)
 *
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * API version: 1.0.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package management_api

type ProjectResponse struct {
	// Id of your project
	Id string `json:"id"`
	// Slug of your organization
	OrganizationId string `json:"organization_id"`
	// Name of your project
	Name string `json:"name"`
	// Region of your project
	Region string `json:"region"`
	// Creation timestamp
	CreatedAt string            `json:"created_at"`
	Database  *DatabaseResponse `json:"database,omitempty"`
}