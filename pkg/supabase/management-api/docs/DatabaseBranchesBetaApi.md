# {{classname}}

All URIs are relative to */*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateBranch**](DatabaseBranchesBetaApi.md#CreateBranch) | **Post** /v1/projects/{ref}/branches | Create a database branch
[**DeleteBranch**](DatabaseBranchesBetaApi.md#DeleteBranch) | **Delete** /v1/branches/{branch_id} | Delete a database branch
[**DisableBranch**](DatabaseBranchesBetaApi.md#DisableBranch) | **Delete** /v1/projects/{ref}/branches | Disables preview branching
[**GetBranchDetails**](DatabaseBranchesBetaApi.md#GetBranchDetails) | **Get** /v1/branches/{branch_id} | Get database branch config
[**GetBranches**](DatabaseBranchesBetaApi.md#GetBranches) | **Get** /v1/projects/{ref}/branches | List all database branches
[**UpdateBranch**](DatabaseBranchesBetaApi.md#UpdateBranch) | **Patch** /v1/branches/{branch_id} | Update database branch config

# **CreateBranch**
> BranchResponse CreateBranch(ctx, body, ref)
Create a database branch

Creates a database branch from the specified project.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**CreateBranchBody**](CreateBranchBody.md)|  | 
  **ref** | **string**| Project ref | 

### Return type

[**BranchResponse**](BranchResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteBranch**
> DeleteBranch(ctx, branchId)
Delete a database branch

Deletes the specified database branch

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **branchId** | **string**| Branch ID | 

### Return type

 (empty response body)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DisableBranch**
> DisableBranch(ctx, ref)
Disables preview branching

Disables preview branching for the specified project

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

# **GetBranchDetails**
> BranchDetailResponse GetBranchDetails(ctx, branchId)
Get database branch config

Fetches configurations of the specified database branch

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **branchId** | **string**| Branch ID | 

### Return type

[**BranchDetailResponse**](BranchDetailResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetBranches**
> []BranchResponse GetBranches(ctx, ref)
List all database branches

Returns all database branches of the specified project.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 

### Return type

[**[]BranchResponse**](BranchResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateBranch**
> BranchResponse UpdateBranch(ctx, body, branchId)
Update database branch config

Updates the configuration of the specified database branch

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**UpdateBranchBody**](UpdateBranchBody.md)|  | 
  **branchId** | **string**| Branch ID | 

### Return type

[**BranchResponse**](BranchResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

