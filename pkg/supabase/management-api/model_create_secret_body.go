/*
 * Supabase API (v1)
 *
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * API version: 1.0.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package management_api

type CreateSecretBody struct {
	// Secret name must not start with the SUPABASE_ prefix.
	Name  string `json:"name"`
	Value string `json:"value"`
}