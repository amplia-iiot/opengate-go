package validator

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

func ValidateUserApiKey(ctx context.Context, db *sql.DB, apiKey string) (userId string, err error) {
	rows, err := db.QueryContext(ctx, "SELECT user_id FROM T_OG_USER WHERE USER_UUID = $1 AND REC_STATUS = 0", apiKey)
	if err != nil {
		return "", fmt.Errorf("error executing query for ApiKey %s: %w", apiKey, err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&userId); err != nil {
			return "", fmt.Errorf("error getting ID for ApiKey %s: %w", apiKey, err)
		}
	} else {
		return "", fmt.Errorf("no validation found for ApiKey %q", apiKey)
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("error on row for ApiKey %q: %w", apiKey, err)
	}

	return userId, nil
}

// ValidateUserHeaders validates the that user contained in the headers exists.
// In the case of a jwt it's validated with the privateKey.
//
// Deprecated: ValidateUserHeaders exists for historical compatibility and
// should not be used for performance reasons. To validate the user use
// ValidateUserHeadersWithKeys or ValidateUserHeadersWithKeyfunc, passing your
// desired implementation. There are keys implementations for the previous mode
// (NewRS256OrHS256Key) or a configurable one for all HS and RS algorithms
// (NewKeys).
func ValidateUserHeaders(ctx context.Context, db *sql.DB, headers http.Header, privateKey string) (userId string, err error) {
	return ValidateUserHeadersWithKeys(ctx, db, headers, NewRS256OrHS256Key(privateKey))
}

// ValidateUserHeadersWithKeys validates that the user contained in the headers
// exists. In the case of a jwt it's validated with the proper key based on the
// algorithm.
func ValidateUserHeadersWithKeys(ctx context.Context, db *sql.DB, headers http.Header, keys *keys) (userId string, err error) {
	return ValidateUserHeadersWithKeyfunc(ctx, db, headers, keys.Keyfunc)
}

// ValidateUserHeadersWithKeyfunc validates that the user contained in the
// headers exists. In the case of a jwt it's validated with the key returned
// by keyfunc.
func ValidateUserHeadersWithKeyfunc(ctx context.Context, db *sql.DB, headers http.Header, keyfunc jwt.Keyfunc) (userId string, err error) {
	apiKey, err := GetApiKeyWithKeyfunc(headers, keyfunc)

	if err != nil {
		return "", err
	}
	return ValidateUserApiKey(ctx, db, apiKey)
}
