/*
 * Supabase API (v1)
 *
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * API version: 1.0.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package management_api

type AllOfDatabaseUpgradeStatusResponseDatabaseUpgradeStatus struct {
	InitiatedAt   string  `json:"initiated_at"`
	TargetVersion float64 `json:"target_version"`
	Error_        string  `json:"error,omitempty"`
	Progress      string  `json:"progress,omitempty"`
	Status        float64 `json:"status"`
}