# {{classname}}

All URIs are relative to */*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetSslEnforcementConfig**](SslEnforcementBetaApi.md#GetSslEnforcementConfig) | **Get** /v1/projects/{ref}/ssl-enforcement | Get project&#x27;s SSL enforcement configuration.
[**UpdateSslEnforcementConfig**](SslEnforcementBetaApi.md#UpdateSslEnforcementConfig) | **Put** /v1/projects/{ref}/ssl-enforcement | Update project&#x27;s SSL enforcement configuration.

# **GetSslEnforcementConfig**
> SslEnforcementResponse GetSslEnforcementConfig(ctx, ref)
Get project's SSL enforcement configuration.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 

### Return type

[**SslEnforcementResponse**](SslEnforcementResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateSslEnforcementConfig**
> SslEnforcementResponse UpdateSslEnforcementConfig(ctx, body, ref)
Update project's SSL enforcement configuration.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**SslEnforcementRequest**](SslEnforcementRequest.md)|  | 
  **ref** | **string**| Project ref | 

### Return type

[**SslEnforcementResponse**](SslEnforcementResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

