# {{classname}}

All URIs are relative to */*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetPgsodiumConfig**](PgsodiumBetaApi.md#GetPgsodiumConfig) | **Get** /v1/projects/{ref}/pgsodium | Gets project&#x27;s pgsodium config
[**UpdatePgsodiumConfig**](PgsodiumBetaApi.md#UpdatePgsodiumConfig) | **Put** /v1/projects/{ref}/pgsodium | Updates project&#x27;s pgsodium config. Updating the root_key can cause all data encrypted with the older key to become inaccessible.

# **GetPgsodiumConfig**
> PgsodiumConfigResponse GetPgsodiumConfig(ctx, ref)
Gets project's pgsodium config

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 

### Return type

[**PgsodiumConfigResponse**](PgsodiumConfigResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdatePgsodiumConfig**
> PgsodiumConfigResponse UpdatePgsodiumConfig(ctx, body, ref)
Updates project's pgsodium config. Updating the root_key can cause all data encrypted with the older key to become inaccessible.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**UpdatePgsodiumConfigBody**](UpdatePgsodiumConfigBody.md)|  | 
  **ref** | **string**| Project ref | 

### Return type

[**PgsodiumConfigResponse**](PgsodiumConfigResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

