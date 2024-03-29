/*
 * Supabase API (v1)
 *
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * API version: 1.0.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package management_api

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/antihax/optional"
)

// Linger please
var (
	_ context.Context
)

type OauthBetaApiService service

/*
OauthBetaApiService Authorize user through oauth
 * @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 * @param clientId
 * @param responseType
 * @param redirectUri
 * @param optional nil or *OauthBetaApiAuthorizeOpts - Optional Parameters:
     * @param "Scope" (optional.String) -
     * @param "State" (optional.String) -
     * @param "ResponseMode" (optional.String) -
     * @param "CodeChallenge" (optional.String) -
     * @param "CodeChallengeMethod" (optional.String) -

*/

type OauthBetaApiAuthorizeOpts struct {
	Scope               optional.String
	State               optional.String
	ResponseMode        optional.String
	CodeChallenge       optional.String
	CodeChallengeMethod optional.String
}

func (a *OauthBetaApiService) Authorize(ctx context.Context, clientId string, responseType string, redirectUri string, localVarOptionals *OauthBetaApiAuthorizeOpts) (*http.Response, error) {
	var (
		localVarHttpMethod = strings.ToUpper("Get")
		localVarPostBody   interface{}
		localVarFileName   string
		localVarFileBytes  []byte
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/v1/oauth/authorize"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	localVarQueryParams.Add("client_id", parameterToString(clientId, ""))
	localVarQueryParams.Add("response_type", parameterToString(responseType, ""))
	localVarQueryParams.Add("redirect_uri", parameterToString(redirectUri, ""))
	if localVarOptionals != nil && localVarOptionals.Scope.IsSet() {
		localVarQueryParams.Add("scope", parameterToString(localVarOptionals.Scope.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.State.IsSet() {
		localVarQueryParams.Add("state", parameterToString(localVarOptionals.State.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.ResponseMode.IsSet() {
		localVarQueryParams.Add("response_mode", parameterToString(localVarOptionals.ResponseMode.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.CodeChallenge.IsSet() {
		localVarQueryParams.Add("code_challenge", parameterToString(localVarOptionals.CodeChallenge.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.CodeChallengeMethod.IsSet() {
		localVarQueryParams.Add("code_challenge_method", parameterToString(localVarOptionals.CodeChallengeMethod.Value(), ""))
	}
	// to determine the Content-Type header
	localVarHttpContentTypes := []string{}

	// set Content-Type header
	localVarHttpContentType := selectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}

	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{}

	// set Accept header
	localVarHttpHeaderAccept := selectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return nil, err
	}

	localVarHttpResponse, err := a.client.callAPI(r)
	if err != nil || localVarHttpResponse == nil {
		return localVarHttpResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHttpResponse.Body)
	localVarHttpResponse.Body.Close()
	if err != nil {
		return localVarHttpResponse, err
	}

	if localVarHttpResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHttpResponse.Status,
		}
		return localVarHttpResponse, newErr
	}

	return localVarHttpResponse, nil
}

/*
OauthBetaApiService Exchange auth code for user&#x27;s access and refresh token
  - @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
  - @param grantType
  - @param clientId
  - @param clientSecret
  - @param code
  - @param codeVerifier
  - @param redirectUri
  - @param refreshToken

@return OAuthTokenResponse
*/
func (a *OauthBetaApiService) Token(ctx context.Context, grantType string, clientId string, clientSecret string, code string, codeVerifier string, redirectUri string, refreshToken string) (OAuthTokenResponse, *http.Response, error) {
	var (
		localVarHttpMethod  = strings.ToUpper("Post")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue OAuthTokenResponse
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/v1/oauth/token"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHttpContentTypes := []string{"application/x-www-form-urlencoded"}

	// set Content-Type header
	localVarHttpContentType := selectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}

	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHttpHeaderAccept := selectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	localVarFormParams.Add("grant_type", parameterToString(grantType, ""))
	localVarFormParams.Add("client_id", parameterToString(clientId, ""))
	localVarFormParams.Add("client_secret", parameterToString(clientSecret, ""))
	localVarFormParams.Add("code", parameterToString(code, ""))
	localVarFormParams.Add("code_verifier", parameterToString(codeVerifier, ""))
	localVarFormParams.Add("redirect_uri", parameterToString(redirectUri, ""))
	localVarFormParams.Add("refresh_token", parameterToString(refreshToken, ""))
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHttpResponse, err := a.client.callAPI(r)
	if err != nil || localVarHttpResponse == nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHttpResponse.Body)
	localVarHttpResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	if localVarHttpResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHttpResponse, err
		}
	}

	if localVarHttpResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHttpResponse.Status,
		}
		if localVarHttpResponse.StatusCode == 201 {
			var v OAuthTokenResponse
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		return localVarReturnValue, localVarHttpResponse, newErr
	}

	return localVarReturnValue, localVarHttpResponse, nil
}