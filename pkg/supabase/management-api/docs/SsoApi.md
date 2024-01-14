# {{classname}}

All URIs are relative to */*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateProviderForProject**](SsoApi.md#CreateProviderForProject) | **Post** /v1/projects/{ref}/config/auth/sso/providers | Creates a new SSO provider
[**GetProviderById**](SsoApi.md#GetProviderById) | **Get** /v1/projects/{ref}/config/auth/sso/providers/{provider_id} | Gets a SSO provider by its UUID
[**ListAllProviders**](SsoApi.md#ListAllProviders) | **Get** /v1/projects/{ref}/config/auth/sso/providers | Lists all SSO providers
[**RemoveProviderById**](SsoApi.md#RemoveProviderById) | **Delete** /v1/projects/{ref}/config/auth/sso/providers/{provider_id} | Removes a SSO provider by its UUID
[**UpdateProviderById**](SsoApi.md#UpdateProviderById) | **Put** /v1/projects/{ref}/config/auth/sso/providers/{provider_id} | Updates a SSO provider by its UUID

# **CreateProviderForProject**
> CreateProviderResponse CreateProviderForProject(ctx, body, ref)
Creates a new SSO provider

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**CreateProviderBody**](CreateProviderBody.md)|  | 
  **ref** | **string**| Project ref | 

### Return type

[**CreateProviderResponse**](CreateProviderResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetProviderById**
> GetProviderResponse GetProviderById(ctx, ref, providerId)
Gets a SSO provider by its UUID

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 
  **providerId** | **string**|  | 

### Return type

[**GetProviderResponse**](GetProviderResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **ListAllProviders**
> ListProvidersResponse ListAllProviders(ctx, ref)
Lists all SSO providers

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 

### Return type

[**ListProvidersResponse**](ListProvidersResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **RemoveProviderById**
> DeleteProviderResponse RemoveProviderById(ctx, ref, providerId)
Removes a SSO provider by its UUID

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 
  **providerId** | **string**|  | 

### Return type

[**DeleteProviderResponse**](DeleteProviderResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateProviderById**
> UpdateProviderResponse UpdateProviderById(ctx, body, ref, providerId)
Updates a SSO provider by its UUID

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**UpdateProviderBody**](UpdateProviderBody.md)|  | 
  **ref** | **string**| Project ref | 
  **providerId** | **string**|  | 

### Return type

[**UpdateProviderResponse**](UpdateProviderResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

