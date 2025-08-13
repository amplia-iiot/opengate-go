package validator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/renanbastos93/fastpath"
)

type ResponseHandler func(c *gin.Context, statusCode int, err error)

type middlewareOptions struct {
	patterns        []fastpath.Path
	ignoredPatterns []fastpath.Path
	allowedProfiles []string
	keyfunc         jwt.Keyfunc
	responseHandler ResponseHandler
}

type middlewareOption func(o *middlewareOptions)

const (
	keyOrganizationId string = "validator:middleware:organization-id" // Can't create custom type to avoid collisions because gin uses string instead of any as keys
)

func OrganizationIdFromContext(ctx context.Context) (organizationId int64, ok bool) {
	organizationId, ok = ctx.Value(keyOrganizationId).(int64)
	return
}

func getMiddlewareOpts(opts ...middlewareOption) (options middlewareOptions) {
	options.patterns = []fastpath.Path{fastpath.New("/organization/:org/*")}
	options.ignoredPatterns = []fastpath.Path{}
	options.keyfunc = func(t *jwt.Token) (interface{}, error) {
		return nil, unsupportedToken(t)
	}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

func UserApiKeyInOrganizationMiddleware(db *sql.DB, opts ...middlewareOption) gin.HandlerFunc {
	return userApiKeyInOrganizationMiddleware(
		getMiddlewareOpts(opts...),
		func(ctx context.Context, organizationName string, headers http.Header, keyfunc jwt.Keyfunc) (organizationId int64, userId int64, profileName string, err error) {
			return ValidateOrganizationHeadersWithKeyfunc(ctx, db, organizationName, headers, keyfunc)
		},
	)
}

type middlewareValidateFunc func(ctx context.Context, organizationName string, headers http.Header, keyfunc jwt.Keyfunc) (organizationId, userId int64, profileName string, err error)

func userApiKeyInOrganizationMiddleware(options middlewareOptions, validate middlewareValidateFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		org := getOrganizationNameFromPaths(options.patterns, path)
		ignored := doesPathMatch(options.ignoredPatterns, path)
		if org == "" && !ignored {
			handleError(c, options, http.StatusNotFound, nil)
			return
		} else if org != "" {
			orgId, _, profile, err := validate(c, org, c.Request.Header, options.keyfunc)

			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					handleError(c, options, http.StatusNotFound, nil)
				} else {
					handleError(c, options, http.StatusUnauthorized, err)
				}
				return
			}

			if len(options.allowedProfiles) > 0 && !slices.Contains(options.allowedProfiles, profile) {
				handleError(c, options, http.StatusUnauthorized, fmt.Errorf("profile %s not allowed", profile))
			}

			c.Set(keyOrganizationId, orgId)
		}
		c.Next()
	}
}

// ValidatedPaths configures the paths that contain an organization name
//
// The position of the organization name in the path is established using
// the ':name' sequence (or any other segment starting with :). Only one
// segment can start with :.
//
// Trailing wildcards are supported with *, but leading wildcards are not supported.
func ValidatePaths(paths ...string) middlewareOption {
	return func(o *middlewareOptions) {
		o.patterns = []fastpath.Path{}
		for _, path := range paths {
			o.patterns = append(o.patterns, fastpath.New(path))
		}
	}
}

// IgnorePaths configures the middleware to not perform the validation in the case that the
// request path matches an ignored path.
//
// Trailing wildcards are supported with *, but leading wildcards are not supported.
func IgnorePaths(paths ...string) middlewareOption {
	return func(o *middlewareOptions) {
		o.ignoredPatterns = []fastpath.Path{}
		for _, path := range paths {
			o.ignoredPatterns = append(o.ignoredPatterns, fastpath.New(path))
		}
	}
}

// Strict configures the middleware to fail validation if the request path does not
// match a validate path and the organization name can't be extracted.
func Strict() middlewareOption {
	return IgnorePaths()
}

// Permissive configures the middleware to skip validation if the request path does not
// match a validate path and the organization name can't be extracted.
func Permissive() middlewareOption {
	return IgnorePaths("/*")
}

// Keyfunc configures the middleware to use the keyfunc to decrypt and validate
// JWTs.
func Keyfunc(keyfunc jwt.Keyfunc) middlewareOption {
	return func(o *middlewareOptions) {
		o.keyfunc = keyfunc
	}
}

// Keys configures the middleware to use the proper key depending on the JWT
// signing algorithm to decrypt and validate JWTs. This method will panic if
// any key is invalid.
func Keys(keys KeysConfig) middlewareOption {
	return Keyfunc(MustNewKeys(keys).Keyfunc)
}

// PrivateKey configures the middleware to use the HS256 private key or RS256
// public key to decrypt and validate JWTs.
func PrivateKey(key string) middlewareOption {
	return Keyfunc(NewRS256OrHS256Key(key).Keyfunc)
}

func AllowedProfiles(profiles ...string) middlewareOption {
	return func(o *middlewareOptions) {
		o.allowedProfiles = slices.Clone(profiles)
	}
}

func ResponseHandlerFunc(responseHandler ResponseHandler) middlewareOption {
	return func(o *middlewareOptions) {
		o.responseHandler = responseHandler
	}
}

func doesPathMatch(patterns []fastpath.Path, path string) bool {
	for _, pattern := range patterns {
		_, ok := pattern.Match(path)
		if ok {
			return true
		}
	}
	return false
}

func getOrganizationNameFromPaths(patterns []fastpath.Path, path string) string {
	for _, pattern := range patterns {
		name := getOrgnizationNameFromPath(pattern, path)
		if name != "" {
			return name
		}
	}
	return ""
}

func getOrgnizationNameFromPath(pattern fastpath.Path, path string) string {
	params, ok := pattern.Match(path)
	if !ok {
		return ""
	}
	return params[0]
}

func handleError(c *gin.Context, options middlewareOptions, statusCode int, err error) {
	if options.responseHandler != nil {
		options.responseHandler(c, statusCode, err)
	} else {
		if err != nil {
			c.AbortWithError(statusCode, err)
		} else {
			c.AbortWithStatus(statusCode)
		}
	}
}
