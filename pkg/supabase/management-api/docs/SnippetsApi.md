# {{classname}}

All URIs are relative to */*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetSnippet**](SnippetsApi.md#GetSnippet) | **Get** /v1/snippets/{id} | Gets a specific SQL snippet
[**ListSnippets**](SnippetsApi.md#ListSnippets) | **Get** /v1/snippets | Lists SQL snippets for the logged in user

# **GetSnippet**
> SnippetResponse GetSnippet(ctx, id)
Gets a specific SQL snippet

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**|  | 

### Return type

[**SnippetResponse**](SnippetResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **ListSnippets**
> SnippetList ListSnippets(ctx, optional)
Lists SQL snippets for the logged in user

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***SnippetsApiListSnippetsOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a SnippetsApiListSnippetsOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **projectRef** | **optional.String**|  | 

### Return type

[**SnippetList**](SnippetList.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

