package validator

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/renanbastos93/fastpath"
)

func TestGetOrganizationNameFromPath(t *testing.T) {
	for name, i := range map[string]struct {
		pattern              string
		path                 string
		expectedOrganization string
	}{
		"starting with slash": {
			pattern:              "/organization/:name",
			path:                 "/organization/org",
			expectedOrganization: "org",
		},
		"starting without slash": {
			pattern:              "organization/:name",
			path:                 "organization/org",
			expectedOrganization: "",
		},
		"trailing wildcard": {
			pattern:              "/organization/:name/*",
			path:                 "/organization/org/component",
			expectedOrganization: "org",
		},
		"trailing wildcard (two segments)": {
			pattern:              "/organization/:name/*",
			path:                 "/organization/org/component/subcomponent",
			expectedOrganization: "org",
		},
		"trailing wildcard (no extra in path)": {
			pattern:              "/organization/:name/*",
			path:                 "/organization/org",
			expectedOrganization: "org",
		},
	} {
		t.Run(name, func(t *testing.T) {
			result := getOrgnizationNameFromPath(fastpath.New(i.pattern), i.path)
			if result != i.expectedOrganization {
				t.Fail()
			}
		})
	}
}

func TestUserApiKeyInOrganizationMiddleware(t *testing.T) {
	for name, i := range map[string]struct {
		opts           []middlewareOption
		validateFunc   middlewareValidateFunc
		requestPath    string
		requestHeader  http.Header
		expectedStatus int
	}{
		"ok with apiKey": {
			opts: []middlewareOption{
				ValidatePaths("/organization/:org"),
			},
			validateFunc:   validateWithApiKey("name", "valid"),
			requestPath:    "/organization/name",
			requestHeader:  withApiKey("valid"),
			expectedStatus: http.StatusOK,
		},
		"ok with wildcard in path": {
			opts: []middlewareOption{
				ValidatePaths("/organization/:org/*"),
			},
			validateFunc:   validateWithApiKey("name", "valid"),
			requestPath:    "/organization/name/more/in/path",
			requestHeader:  withApiKey("valid"),
			expectedStatus: http.StatusOK,
		},
		"unauthorized without apiKey": {
			opts: []middlewareOption{
				ValidatePaths("/organization/:org"),
			},
			validateFunc:   validateWithApiKey("name", "valid"),
			requestPath:    "/organization/name",
			expectedStatus: http.StatusUnauthorized,
		},
		"forbidden with invalid apiKey": {
			opts: []middlewareOption{
				ValidatePaths("/organization/:org"),
			},
			validateFunc:   validateWithApiKey("name", "valid"),
			requestPath:    "/organization/name",
			requestHeader:  withApiKey("invalid"),
			expectedStatus: http.StatusUnauthorized,
		},
		"forbidden with invalid apiKey in permissive mode": {
			opts: []middlewareOption{
				ValidatePaths("/organization/:org"),
				Permissive(),
			},
			validateFunc:   validateWithApiKey("name", "valid"),
			requestPath:    "/organization/name",
			requestHeader:  withApiKey("invalid"),
			expectedStatus: http.StatusUnauthorized,
		},
		"ok with invalid apiKey in permissive mode in unmatched path": {
			opts: []middlewareOption{
				ValidatePaths("/organization/:org"),
				Permissive(),
			},
			validateFunc:   failIfValidateCalled(t),
			requestPath:    "/other",
			requestHeader:  withApiKey("invalid"),
			expectedStatus: http.StatusOK,
		},
		"ok without apiKey in permissive mode in unmatched path": {
			opts: []middlewareOption{
				ValidatePaths("/organization/:org"),
				Permissive(),
			},
			validateFunc:   failIfValidateCalled(t),
			requestPath:    "/other",
			expectedStatus: http.StatusOK,
		},
		"forbidden with invalid apiKey in strict mode": {
			opts: []middlewareOption{
				ValidatePaths("/organization/:org"),
				Strict(),
			},
			validateFunc:   validateWithApiKey("name", "valid"),
			requestPath:    "/organization/name",
			requestHeader:  withApiKey("invalid"),
			expectedStatus: http.StatusUnauthorized,
		},
		"unauthorized without apiKey in strict mode": {
			opts: []middlewareOption{
				ValidatePaths("/organization/:org"),
				Strict(),
			},
			validateFunc:   validateWithApiKey("name", "valid"),
			requestPath:    "/organization/name",
			expectedStatus: http.StatusUnauthorized,
		},
		"not found in strict mode in unmatched path": {
			opts: []middlewareOption{
				ValidatePaths("/organization/:org"),
				Strict(),
			},
			validateFunc:   failIfValidateCalled(t),
			requestPath:    "/other",
			expectedStatus: http.StatusNotFound,
		},
		"ok skipping path": {
			opts: []middlewareOption{
				ValidatePaths("/organization/:org"),
				IgnorePaths("/skip"),
			},
			validateFunc:   failIfValidateCalled(t),
			requestPath:    "/skip",
			expectedStatus: http.StatusOK,
		},
		"not found unsupported path": {
			opts: []middlewareOption{
				ValidatePaths("/organization/:org"),
			},
			validateFunc:   validateWithApiKey("name", "valid"),
			requestPath:    "/other",
			expectedStatus: http.StatusNotFound,
		},
		"ok not checked profile": {
			opts: []middlewareOption{
				ValidatePaths("/organization/:org"),
			},
			validateFunc:   validateWithApiKeyReturnProfile("name", "valid", "not allowed"),
			requestPath:    "/organization/name",
			requestHeader:  withApiKey("valid"),
			expectedStatus: http.StatusOK,
		},
		"ok checking profile": {
			opts: []middlewareOption{
				ValidatePaths("/organization/:org"),
				AllowedProfiles("allowed"),
			},
			validateFunc:   validateWithApiKeyReturnProfile("name", "valid", "allowed"),
			requestPath:    "/organization/name",
			requestHeader:  withApiKey("valid"),
			expectedStatus: http.StatusOK,
		},
		"forbidden not allowed profile": {
			opts: []middlewareOption{
				ValidatePaths("/organization/:org"),
				AllowedProfiles("allowed"),
			},
			validateFunc:   validateWithApiKeyReturnProfile("name", "valid", "not allowed"),
			requestPath:    "/organization/name",
			requestHeader:  withApiKey("valid"),
			expectedStatus: http.StatusUnauthorized,
		},
	} {
		gin.SetMode(gin.TestMode)
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, engine := gin.CreateTestContext(w)
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, i.requestPath, nil)
			req.Header = i.requestHeader

			engine.Use(userApiKeyInOrganizationMiddleware(getMiddlewareOpts(i.opts...), i.validateFunc))
			engine.GET(i.requestPath, func(c *gin.Context) {
				c.String(http.StatusOK, "handler reached")
			})
			engine.ServeHTTP(w, req)

			if i.expectedStatus != w.Result().StatusCode {
				t.Errorf("status code: %d (expected %d)", w.Result().StatusCode, i.expectedStatus)
			}
		})
	}
}

func withApiKey(key string) http.Header {
	header := make(http.Header)
	header.Set(HeaderApiKey, key)
	return header
}

func validateWithApiKeyReturnProfile(validOrg, validApiKey, profileToReturn string) func(ctx context.Context, organizationName string, headers http.Header, keyfunc jwt.Keyfunc) (organizationId int64, userId int64, profileName string, err error) {
	return func(ctx context.Context, organizationName string, headers http.Header, keyfunc jwt.Keyfunc) (organizationId int64, userId int64, profileName string, err error) {
		apiKey := headers.Get(HeaderApiKey)
		if apiKey == "" {
			return 0, 0, "", ErrNoAuth
		}
		if organizationName == validOrg && apiKey == validApiKey {
			return 0, 0, profileToReturn, nil
		}
		return 0, 0, "", fmt.Errorf("invalid org %s and api key %s", organizationName, apiKey)
	}
}

func validateWithApiKey(validOrg, validApiKey string) func(ctx context.Context, organizationName string, headers http.Header, keyfunc jwt.Keyfunc) (organizationId int64, userId int64, profileName string, err error) {
	return validateWithApiKeyReturnProfile(validOrg, validApiKey, "")
}

func failIfValidateCalled(t *testing.T) func(ctx context.Context, organizationName string, headers http.Header, keyfunc jwt.Keyfunc) (organizationId int64, userId int64, profileName string, err error) {
	return func(ctx context.Context, organizationName string, headers http.Header, keyfunc jwt.Keyfunc) (organizationId int64, userId int64, profileName string, err error) {
		t.Error("unwanted validate called")
		return
	}
}
