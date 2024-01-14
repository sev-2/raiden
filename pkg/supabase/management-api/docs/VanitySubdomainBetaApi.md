# {{classname}}

All URIs are relative to */*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ActivateVanitySubdomainPlease**](VanitySubdomainBetaApi.md#ActivateVanitySubdomainPlease) | **Post** /v1/projects/{ref}/vanity-subdomain/activate | Activates a vanity subdomain for a project.
[**CheckVanitySubdomainAvailability**](VanitySubdomainBetaApi.md#CheckVanitySubdomainAvailability) | **Post** /v1/projects/{ref}/vanity-subdomain/check-availability | Checks vanity subdomain availability
[**GetVanitySubdomainConfig**](VanitySubdomainBetaApi.md#GetVanitySubdomainConfig) | **Get** /v1/projects/{ref}/vanity-subdomain | Gets current vanity subdomain config
[**RemoveVanitySubdomainConfig**](VanitySubdomainBetaApi.md#RemoveVanitySubdomainConfig) | **Delete** /v1/projects/{ref}/vanity-subdomain | Deletes a project&#x27;s vanity subdomain configuration

# **ActivateVanitySubdomainPlease**
> ActivateVanitySubdomainResponse ActivateVanitySubdomainPlease(ctx, body, ref)
Activates a vanity subdomain for a project.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**VanitySubdomainBody**](VanitySubdomainBody.md)|  | 
  **ref** | **string**| Project ref | 

### Return type

[**ActivateVanitySubdomainResponse**](ActivateVanitySubdomainResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **CheckVanitySubdomainAvailability**
> SubdomainAvailabilityResponse CheckVanitySubdomainAvailability(ctx, body, ref)
Checks vanity subdomain availability

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**VanitySubdomainBody**](VanitySubdomainBody.md)|  | 
  **ref** | **string**| Project ref | 

### Return type

[**SubdomainAvailabilityResponse**](SubdomainAvailabilityResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetVanitySubdomainConfig**
> VanitySubdomainConfigResponse GetVanitySubdomainConfig(ctx, ref)
Gets current vanity subdomain config

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 

### Return type

[**VanitySubdomainConfigResponse**](VanitySubdomainConfigResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **RemoveVanitySubdomainConfig**
> RemoveVanitySubdomainConfig(ctx, ref)
Deletes a project's vanity subdomain configuration

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

