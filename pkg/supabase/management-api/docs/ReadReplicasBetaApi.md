# {{classname}}

All URIs are relative to */*

Method | HTTP request | Description
------------- | ------------- | -------------
[**RemoveReadReplica**](ReadReplicasBetaApi.md#RemoveReadReplica) | **Post** /v1/projects/{ref}/read-replicas/remove | Remove a read replica
[**SetUpReadReplica**](ReadReplicasBetaApi.md#SetUpReadReplica) | **Post** /v1/projects/{ref}/read-replicas/setup | Set up a read replica

# **RemoveReadReplica**
> RemoveReadReplica(ctx, body, ref)
Remove a read replica

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**RemoveReadReplicaBody**](RemoveReadReplicaBody.md)|  | 
  **ref** | **string**| Project ref | 

### Return type

 (empty response body)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **SetUpReadReplica**
> SetUpReadReplica(ctx, body, ref)
Set up a read replica

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**SetUpReadReplicaBody**](SetUpReadReplicaBody.md)|  | 
  **ref** | **string**| Project ref | 

### Return type

 (empty response body)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

