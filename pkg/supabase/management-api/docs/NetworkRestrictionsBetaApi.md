# {{classname}}

All URIs are relative to */*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ApplyNetworkRestrictions**](NetworkRestrictionsBetaApi.md#ApplyNetworkRestrictions) | **Post** /v1/projects/{ref}/network-restrictions/apply | Updates project&#x27;s network restrictions
[**GetNetworkRestrictions**](NetworkRestrictionsBetaApi.md#GetNetworkRestrictions) | **Get** /v1/projects/{ref}/network-restrictions | Gets project&#x27;s network restrictions

# **ApplyNetworkRestrictions**
> NetworkRestrictionsResponse ApplyNetworkRestrictions(ctx, body, ref)
Updates project's network restrictions

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**NetworkRestrictionsRequest**](NetworkRestrictionsRequest.md)|  | 
  **ref** | **string**| Project ref | 

### Return type

[**NetworkRestrictionsResponse**](NetworkRestrictionsResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetNetworkRestrictions**
> NetworkRestrictionsResponse GetNetworkRestrictions(ctx, ref)
Gets project's network restrictions

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 

### Return type

[**NetworkRestrictionsResponse**](NetworkRestrictionsResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

