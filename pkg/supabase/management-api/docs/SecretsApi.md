# {{classname}}

All URIs are relative to */*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateSecrets**](SecretsApi.md#CreateSecrets) | **Post** /v1/projects/{ref}/secrets | Bulk create secrets
[**DeleteSecrets**](SecretsApi.md#DeleteSecrets) | **Delete** /v1/projects/{ref}/secrets | Bulk delete secrets
[**GetSecrets**](SecretsApi.md#GetSecrets) | **Get** /v1/projects/{ref}/secrets | List all secrets

# **CreateSecrets**
> CreateSecrets(ctx, body, ref)
Bulk create secrets

Creates multiple secrets and adds them to the specified project.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**[]CreateSecretBody**](CreateSecretBody.md)|  | 
  **ref** | **string**| Project ref | 

### Return type

 (empty response body)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteSecrets**
> interface{} DeleteSecrets(ctx, body, ref)
Bulk delete secrets

Deletes all secrets with the given names from the specified project

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**[]string**](string.md)|  | 
  **ref** | **string**| Project ref | 

### Return type

[**interface{}**](interface{}.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetSecrets**
> []SecretResponse GetSecrets(ctx, ref)
List all secrets

Returns all secrets you've previously added to the specified project.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 

### Return type

[**[]SecretResponse**](SecretResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

