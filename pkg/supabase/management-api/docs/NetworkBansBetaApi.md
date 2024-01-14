# {{classname}}

All URIs are relative to */*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetNetworkBans**](NetworkBansBetaApi.md#GetNetworkBans) | **Post** /v1/projects/{ref}/network-bans/retrieve | Gets project&#x27;s network bans
[**RemoveNetworkBan**](NetworkBansBetaApi.md#RemoveNetworkBan) | **Delete** /v1/projects/{ref}/network-bans | Remove network bans.

# **GetNetworkBans**
> NetworkBanResponse GetNetworkBans(ctx, ref)
Gets project's network bans

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 

### Return type

[**NetworkBanResponse**](NetworkBanResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **RemoveNetworkBan**
> RemoveNetworkBan(ctx, body, ref)
Remove network bans.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**RemoveNetworkBanRequest**](RemoveNetworkBanRequest.md)|  | 
  **ref** | **string**| Project ref | 

### Return type

 (empty response body)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

