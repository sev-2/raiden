# {{classname}}

All URIs are relative to */*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetUpgradeStatus**](DatabaseVersionUpgradeBetaApi.md#GetUpgradeStatus) | **Get** /v1/projects/{ref}/upgrade/status | Gets the latest status of the project&#x27;s upgrade
[**UpgradeEligibilityInformation**](DatabaseVersionUpgradeBetaApi.md#UpgradeEligibilityInformation) | **Get** /v1/projects/{ref}/upgrade/eligibility | Returns the project&#x27;s eligibility for upgrades
[**UpgradeProject**](DatabaseVersionUpgradeBetaApi.md#UpgradeProject) | **Post** /v1/projects/{ref}/upgrade | Upgrades the project&#x27;s Postgres version

# **GetUpgradeStatus**
> DatabaseUpgradeStatusResponse GetUpgradeStatus(ctx, ref)
Gets the latest status of the project's upgrade

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 

### Return type

[**DatabaseUpgradeStatusResponse**](DatabaseUpgradeStatusResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpgradeEligibilityInformation**
> ProjectUpgradeEligibilityResponse UpgradeEligibilityInformation(ctx, ref)
Returns the project's eligibility for upgrades

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ref** | **string**| Project ref | 

### Return type

[**ProjectUpgradeEligibilityResponse**](ProjectUpgradeEligibilityResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpgradeProject**
> ProjectUpgradeInitiateResponse UpgradeProject(ctx, body, ref)
Upgrades the project's Postgres version

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**UpgradeDatabaseBody**](UpgradeDatabaseBody.md)|  | 
  **ref** | **string**| Project ref | 

### Return type

[**ProjectUpgradeInitiateResponse**](ProjectUpgradeInitiateResponse.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

