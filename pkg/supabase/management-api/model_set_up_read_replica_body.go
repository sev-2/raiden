/*
 * Supabase API (v1)
 *
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * API version: 1.0.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package management_api

type SetUpReadReplicaBody struct {
	// Region you want your read replica to reside in
	ReadReplicaRegion string `json:"read_replica_region"`
}