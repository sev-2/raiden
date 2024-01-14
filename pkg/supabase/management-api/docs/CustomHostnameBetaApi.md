# {{classname}}

All URIs are relative to */*

Method | HTTP request | Description
------------- | ------------- | -------------
[**Activate**](CustomHostnameBetaApi.md#Activate) | **Post** /v1/projects/{ref}/custom-hostname/activate | Activates a custom hostname for a project.
[**CreateCustomHostnameConfig**](CustomHostnameBetaApi.md#CreateCustomHostnameConfig) | **Post** /v1/projects/{ref}/custom-hostname/initialize | Updates project&#x27;s custom hostname configuration
[**GetCustomHostnameConfig**](CustomHostnameBetaApi.md#GetCustomHostnameConfig) | **Get** /v1/projects/{ref}/custom-hostname | Gets project&#x27;s custom hostname config
[**RemoveCustomHostnameConfig**](CustomHostnameBetaApi.md#RemoveCustomHostnameConfig) | **Delete** /v1/projects/{ref}/custom-hostname | Deletes a project&#x27;s custom hostname configuration
[**Reverify**](CustomHostnameBetaApi.md#Reverify) | **Post** /v1/projects/{ref}/custom-hostname/reverify | Attempts to verify the DNS configuration for project&#x27;s custom hostname configuration

# **Activate**
> UpdateCustomHostnameResponse Activate(ctx, ref)
Activates a custom hostname for a project.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 

### Return type

[**UpdateCustomHostnameResponse**](UpdateCustomHostnameResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **CreateCustomHostnameConfig**
> UpdateCustomHostnameResponse CreateCustomHostnameConfig(ctx, body, ref)
Updates project's custom hostname configuration

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**UpdateCustomHostnameBody**](UpdateCustomHostnameBody.md)|  | 
  **ref** | **string**| Project ref | 

### Return type

[**UpdateCustomHostnameResponse**](UpdateCustomHostnameResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetCustomHostnameConfig**
> UpdateCustomHostnameResponse GetCustomHostnameConfig(ctx, ref)
Gets project's custom hostname config

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 

### Return type

[**UpdateCustomHostnameResponse**](UpdateCustomHostnameResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **RemoveCustomHostnameConfig**
> RemoveCustomHostnameConfig(ctx, ref)
Deletes a project's custom hostname configuration

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 

### Return type

 (empty response body)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **Reverify**
> UpdateCustomHostnameResponse Reverify(ctx, ref)
Attempts to verify the DNS configuration for project's custom hostname configuration

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 

### Return type

[**UpdateCustomHostnameResponse**](UpdateCustomHostnameResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

