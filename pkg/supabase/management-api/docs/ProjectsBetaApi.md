# {{classname}}

All URIs are relative to */*

Method | HTTP request | Description
------------- | ------------- | -------------
[**V1EnableDatabaseWebhooks**](ProjectsBetaApi.md#V1EnableDatabaseWebhooks) | **Post** /v1/projects/{ref}/database/webhooks/enable | Enables Database Webhooks on the project
[**V1RunQuery**](ProjectsBetaApi.md#V1RunQuery) | **Post** /v1/projects/{ref}/database/query | Run sql query

# **V1EnableDatabaseWebhooks**
> V1EnableDatabaseWebhooks(ctx, ref)
Enables Database Webhooks on the project

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

# **V1RunQuery**
> interface{} V1RunQuery(ctx, body, ref)
Run sql query

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**RunQueryBody**](RunQueryBody.md)|  | 
  **ref** | **string**| Project ref | 

### Return type

[**interface{}**](interface{}.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

