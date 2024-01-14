# {{classname}}

All URIs are relative to */*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CheckServiceHealth**](ServicesApi.md#CheckServiceHealth) | **Get** /v1/projects/{ref}/health | Gets project&#x27;s service health status
[**GetPostgRESTConfig**](ServicesApi.md#GetPostgRESTConfig) | **Get** /v1/projects/{ref}/postgrest | Gets project&#x27;s postgrest config
[**UpdatePostgRESTConfig**](ServicesApi.md#UpdatePostgRESTConfig) | **Patch** /v1/projects/{ref}/postgrest | Updates project&#x27;s postgrest config

# **CheckServiceHealth**
> []ServiceHealthResponse CheckServiceHealth(ctx, ref, services, optional)
Gets project's service health status

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 
  **services** | [**[]string**](string.md)|  | 
 **optional** | ***ServicesApiCheckServiceHealthOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ServicesApiCheckServiceHealthOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **timeoutMs** | **optional.Int32**|  | 

### Return type

[**[]ServiceHealthResponse**](ServiceHealthResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetPostgRESTConfig**
> PostgrestConfigWithJwtSecretResponse GetPostgRESTConfig(ctx, ref)
Gets project's postgrest config

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 

### Return type

[**PostgrestConfigWithJwtSecretResponse**](PostgrestConfigWithJWTSecretResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdatePostgRESTConfig**
> PostgrestConfigResponse UpdatePostgRESTConfig(ctx, body, ref)
Updates project's postgrest config

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**UpdatePostgrestConfigBody**](UpdatePostgrestConfigBody.md)|  | 
  **ref** | **string**| Project ref | 

### Return type

[**PostgrestConfigResponse**](PostgrestConfigResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

