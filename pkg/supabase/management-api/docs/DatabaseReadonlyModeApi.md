# {{classname}}

All URIs are relative to */*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetReadOnlyModeStatus**](DatabaseReadonlyModeApi.md#GetReadOnlyModeStatus) | **Get** /v1/projects/{ref}/readonly | Returns project&#x27;s readonly mode status
[**TemporarilyDisableReadonlyMode**](DatabaseReadonlyModeApi.md#TemporarilyDisableReadonlyMode) | **Post** /v1/projects/{ref}/readonly/temporary-disable | Disables project&#x27;s readonly mode for the next 15 minutes

# **GetReadOnlyModeStatus**
> ReadOnlyStatusResponse GetReadOnlyModeStatus(ctx, ref)
Returns project's readonly mode status

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 

### Return type

[**ReadOnlyStatusResponse**](ReadOnlyStatusResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **TemporarilyDisableReadonlyMode**
> TemporarilyDisableReadonlyMode(ctx, ref)
Disables project's readonly mode for the next 15 minutes

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

