# {{classname}}

All URIs are relative to */*

Method | HTTP request | Description
------------- | ------------- | -------------
[**Authorize**](OauthBetaApi.md#Authorize) | **Get** /v1/oauth/authorize | Authorize user through oauth
[**Token**](OauthBetaApi.md#Token) | **Post** /v1/oauth/token | Exchange auth code for user&#x27;s access and refresh token

# **Authorize**
> Authorize(ctx, clientId, responseType, redirectUri, optional)
Authorize user through oauth

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **clientId** | **string**|  | 
  **responseType** | **string**|  | 
  **redirectUri** | **string**|  | 
 **optional** | ***OauthBetaApiAuthorizeOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a OauthBetaApiAuthorizeOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **scope** | **optional.String**|  | 
 **state** | **optional.String**|  | 
 **responseMode** | **optional.String**|  | 
 **codeChallenge** | **optional.String**|  | 
 **codeChallengeMethod** | **optional.String**|  | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **Token**
> OAuthTokenResponse Token(ctx, grantType, clientId, clientSecret, code, codeVerifier, redirectUri, refreshToken)
Exchange auth code for user's access and refresh token

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **grantType** | **string**|  | 
  **clientId** | **string**|  | 
  **clientSecret** | **string**|  | 
  **code** | **string**|  | 
  **codeVerifier** | **string**|  | 
  **redirectUri** | **string**|  | 
  **refreshToken** | **string**|  | 

### Return type

[**OAuthTokenResponse**](OAuthTokenResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/x-www-form-urlencoded
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

