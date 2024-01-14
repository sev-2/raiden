# {{classname}}

All URIs are relative to */*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateProject**](ProjectsApi.md#CreateProject) | **Post** /v1/projects | Create a project
[**DeleteProject**](ProjectsApi.md#DeleteProject) | **Delete** /v1/projects/{ref} | Deletes the given project
[**GetProjectApiKeys**](ProjectsApi.md#GetProjectApiKeys) | **Get** /v1/projects/{ref}/api-keys | Get project api keys
[**GetProjects**](ProjectsApi.md#GetProjects) | **Get** /v1/projects | List all projects
[**GetTypescriptTypes**](ProjectsApi.md#GetTypescriptTypes) | **Get** /v1/projects/{ref}/types/typescript | Generate TypeScript types

# **CreateProject**
> ProjectResponse CreateProject(ctx, body)
Create a project

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**CreateProjectBody**](CreateProjectBody.md)|  | 

### Return type

[**ProjectResponse**](ProjectResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteProject**
> ProjectRefResponse DeleteProject(ctx, ref)
Deletes the given project

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 

### Return type

[**ProjectRefResponse**](ProjectRefResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetProjectApiKeys**
> []ApiKeyResponse GetProjectApiKeys(ctx, ref)
Get project api keys

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 

### Return type

[**[]ApiKeyResponse**](ApiKeyResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetProjects**
> []ProjectResponse GetProjects(ctx, )
List all projects

Returns a list of all projects you've previously created.

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**[]ProjectResponse**](ProjectResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetTypescriptTypes**
> TypescriptResponse GetTypescriptTypes(ctx, ref, optional)
Generate TypeScript types

Returns the TypeScript types of your schema for use with supabase-js.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 
 **optional** | ***ProjectsApiGetTypescriptTypesOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ProjectsApiGetTypescriptTypesOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **includedSchemas** | **optional.String**|  | [default to public]

### Return type

[**TypescriptResponse**](TypescriptResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

