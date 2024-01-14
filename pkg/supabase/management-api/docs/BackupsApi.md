# {{classname}}

All URIs are relative to */*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetBackups**](BackupsApi.md#GetBackups) | **Get** /v1/projects/{ref}/database/backups | Lists all backups
[**V1RestorePitr**](BackupsApi.md#V1RestorePitr) | **Post** /v1/projects/{ref}/database/backups/restore-pitr | Restores a PITR backup for a database

# **GetBackups**
> V1BackupsResponse GetBackups(ctx, ref)
Lists all backups

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 

### Return type

[**V1BackupsResponse**](V1BackupsResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **V1RestorePitr**
> V1RestorePitr(ctx, body, ref)
Restores a PITR backup for a database

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**V1RestorePitrBody**](V1RestorePitrBody.md)|  | 
  **ref** | **string**| Project ref | 

### Return type

 (empty response body)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

