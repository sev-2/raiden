# {{classname}}

All URIs are relative to */*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateFunction**](FunctionsApi.md#CreateFunction) | **Post** /v1/projects/{ref}/functions | Create a function
[**DeleteFunction**](FunctionsApi.md#DeleteFunction) | **Delete** /v1/projects/{ref}/functions/{function_slug} | Delete a function
[**GetFunction**](FunctionsApi.md#GetFunction) | **Get** /v1/projects/{ref}/functions/{function_slug} | Retrieve a function
[**GetFunctionBody**](FunctionsApi.md#GetFunctionBody) | **Get** /v1/projects/{ref}/functions/{function_slug}/body | Retrieve a function body
[**GetFunctions**](FunctionsApi.md#GetFunctions) | **Get** /v1/projects/{ref}/functions | List all functions
[**UpdateFunction**](FunctionsApi.md#UpdateFunction) | **Patch** /v1/projects/{ref}/functions/{function_slug} | Update a function

# **CreateFunction**
> FunctionResponse CreateFunction(ctx, body, ref, optional)
Create a function

Creates a function and adds it to the specified project.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**CreateFunctionBody**](CreateFunctionBody.md)|  | 
  **ref** | **string**| Project ref | 
 **optional** | ***FunctionsApiCreateFunctionOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a FunctionsApiCreateFunctionOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **slug** | **optional.**|  | 
 **name** | **optional.**|  | 
 **verifyJwt** | **optional.**|  | 
 **importMap** | **optional.**|  | 
 **entrypointPath** | **optional.**|  | 
 **importMapPath** | **optional.**|  | 

### Return type

[**FunctionResponse**](FunctionResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json, application/vnd.denoland.eszip
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteFunction**
> DeleteFunction(ctx, ref, functionSlug)
Delete a function

Deletes a function with the specified slug from the specified project.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 
  **functionSlug** | **string**| Function slug | 

### Return type

 (empty response body)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetFunction**
> FunctionSlugResponse GetFunction(ctx, ref, functionSlug)
Retrieve a function

Retrieves a function with the specified slug and project.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 
  **functionSlug** | **string**| Function slug | 

### Return type

[**FunctionSlugResponse**](FunctionSlugResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetFunctionBody**
> GetFunctionBody(ctx, ref, functionSlug)
Retrieve a function body

Retrieves a function body for the specified slug and project.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 
  **functionSlug** | **string**| Function slug | 

### Return type

 (empty response body)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetFunctions**
> []FunctionResponse GetFunctions(ctx, ref)
List all functions

Returns all functions you've previously added to the specified project.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 

### Return type

[**[]FunctionResponse**](FunctionResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateFunction**
> FunctionResponse UpdateFunction(ctx, body, ref, functionSlug, optional)
Update a function

Updates a function with the specified slug and project.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**UpdateFunctionBody**](UpdateFunctionBody.md)|  | 
  **ref** | **string**| Project ref | 
  **functionSlug** | **string**| Function slug | 
 **optional** | ***FunctionsApiUpdateFunctionOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a FunctionsApiUpdateFunctionOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **slug** | **optional.**|  | 
 **name** | **optional.**|  | 
 **verifyJwt** | **optional.**|  | 
 **importMap** | **optional.**|  | 
 **entrypointPath** | **optional.**|  | 
 **importMapPath** | **optional.**|  | 

### Return type

[**FunctionResponse**](FunctionResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json, application/vnd.denoland.eszip
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

