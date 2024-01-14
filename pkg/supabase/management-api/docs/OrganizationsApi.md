# {{classname}}

All URIs are relative to */*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateOrganization**](OrganizationsApi.md#CreateOrganization) | **Post** /v1/organizations | Create an organization
[**GetOrganization**](OrganizationsApi.md#GetOrganization) | **Get** /v1/organizations/{slug} | Gets information about the organization
[**GetOrganizations**](OrganizationsApi.md#GetOrganizations) | **Get** /v1/organizations | List all organizations
[**V1ListOrganizationMembers**](OrganizationsApi.md#V1ListOrganizationMembers) | **Get** /v1/organizations/{slug}/members | List members of an organization

# **CreateOrganization**
> OrganizationResponseV1 CreateOrganization(ctx, body)
Create an organization

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**CreateOrganizationBodyV1**](CreateOrganizationBodyV1.md)|  | 

### Return type

[**OrganizationResponseV1**](OrganizationResponseV1.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetOrganization**
> V1OrganizationSlugResponse GetOrganization(ctx, slug)
Gets information about the organization

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **slug** | **string**|  | 

### Return type

[**V1OrganizationSlugResponse**](V1OrganizationSlugResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetOrganizations**
> []OrganizationResponseV1 GetOrganizations(ctx, )
List all organizations

Returns a list of organizations that you currently belong to.

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**[]OrganizationResponseV1**](OrganizationResponseV1.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **V1ListOrganizationMembers**
> []V1OrganizationMemberResponse V1ListOrganizationMembers(ctx, slug)
List members of an organization

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **slug** | **string**|  | 

### Return type

[**[]V1OrganizationMemberResponse**](V1OrganizationMemberResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

