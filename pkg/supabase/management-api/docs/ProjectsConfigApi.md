# {{classname}}

All URIs are relative to */*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetConfig**](ProjectsConfigApi.md#GetConfig) | **Get** /v1/projects/{ref}/config/database/postgres | Gets project&#x27;s Postgres config
[**GetV1AuthConfig**](ProjectsConfigApi.md#GetV1AuthConfig) | **Get** /v1/projects/{ref}/config/auth | Gets project&#x27;s auth config
[**UpdateConfig**](ProjectsConfigApi.md#UpdateConfig) | **Put** /v1/projects/{ref}/config/database/postgres | Updates project&#x27;s Postgres config
[**UpdateV1AuthConfig**](ProjectsConfigApi.md#UpdateV1AuthConfig) | **Patch** /v1/projects/{ref}/config/auth | Updates a project&#x27;s auth config
[**V1GetPgbouncerConfig**](ProjectsConfigApi.md#V1GetPgbouncerConfig) | **Get** /v1/projects/{ref}/config/database/pgbouncer | Get project&#x27;s pgbouncer config

# **GetConfig**
> PostgresConfigResponse GetConfig(ctx, ref)
Gets project's Postgres config

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 

### Return type

[**PostgresConfigResponse**](PostgresConfigResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetV1AuthConfig**
> AuthConfigResponse GetV1AuthConfig(ctx, ref)
Gets project's auth config

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 

### Return type

[**AuthConfigResponse**](AuthConfigResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateConfig**
> PostgresConfigResponse UpdateConfig(ctx, body, ref)
Updates project's Postgres config

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**UpdatePostgresConfigBody**](UpdatePostgresConfigBody.md)|  | 
  **ref** | **string**| Project ref | 

### Return type

[**PostgresConfigResponse**](PostgresConfigResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateV1AuthConfig**
> AuthConfigResponse UpdateV1AuthConfig(ctx, body, ref)
Updates a project's auth config

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**UpdateAuthConfigBody**](UpdateAuthConfigBody.md)|  | 
  **ref** | **string**| Project ref | 

### Return type

[**AuthConfigResponse**](AuthConfigResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **V1GetPgbouncerConfig**
> V1PgbouncerConfigResponse V1GetPgbouncerConfig(ctx, ref)
Get project's pgbouncer config

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 

### Return type

[**V1PgbouncerConfigResponse**](V1PgbouncerConfigResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

