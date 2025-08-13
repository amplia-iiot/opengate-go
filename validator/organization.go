package validator

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

func ValidateOrganizationApiKey(ctx context.Context, db *sql.DB, organizationName string, apiKey string) (organizationId, userId int64, profileName string, err error) {
	rows, err := db.QueryContext(
		ctx,
		`SELECT org.organization_id, us.user_id, prof.name FROM t_og_organization org
				JOIN t_og_domain_organization dorg ON dorg.organization_id = org.organization_id
				JOIN t_og_user us ON us.domain_id = dorg.domain_id
				join t_og_profile prof on prof.profile_id = us.profile_id
			WHERE us.user_uuid = $1
				AND org.name = $2
				AND org.rec_status = 0
				AND dorg.rec_status = 0
				AND us.rec_status = 0
				AND prof.rec_status = 0`,
		apiKey,
		organizationName,
	)
	if err != nil {
		return 0, 0, "", fmt.Errorf("error executing query for Organization %s and ApiKey %s: %w", organizationName, apiKey, err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&organizationId, &userId, &profileName); err != nil {
			return 0, 0, "", fmt.Errorf("error getting ID for Organization %s and ApiKey %s: %w", organizationName, apiKey, err)
		}
	} else {
		return 0, 0, "", fmt.Errorf("no validation found for Organization %s and ApiKey %s: %w", organizationName, apiKey, sql.ErrNoRows)
	}

	if err := rows.Err(); err != nil {
		return 0, 0, "", fmt.Errorf("error on row for Organization %s and ApiKey %s: %w", organizationName, apiKey, err)
	}

	return organizationId, userId, profileName, nil
}

// ValidateOrganizationHeaders validates that the user contained in
// the headers can operate in the organization. In the case of a jwt it's
// validated with the privateKey.
//
// Deprecated: ValidateOrganizationHeaders exists for historical compatibility
// and should not be used for performance reasons. To validate the user in the
// organization use ValidateOrganizationHeadersWithKeys or
// ValidateOrganizationHeadersWithKeyfunc, passing your desired implementation.
// There are keys implementations for the previous mode (NewRS256OrHS256Key) or
// a configurable one for all HS and RS algorithms (NewKeys).
func ValidateOrganizationHeaders(ctx context.Context, db *sql.DB, organizationName string, headers http.Header, privateKey string) (organizationId, userId int64, profileName string, err error) {
	return ValidateOrganizationHeadersWithKeys(ctx, db, organizationName, headers, NewRS256OrHS256Key(privateKey))
}

// ValidateOrganizationHeadersWithKeys validates that the user contained in
// the headers can operate in the organization. In the case of a jwt it's
// validated with the proper key based on the algorithm.
func ValidateOrganizationHeadersWithKeys(ctx context.Context, db *sql.DB, organizationName string, headers http.Header, keys *keys) (organizationId, userId int64, profileName string, err error) {
	return ValidateOrganizationHeadersWithKeyfunc(ctx, db, organizationName, headers, keys.Keyfunc)
}

// ValidateOrganizationHeadersWithKeyfunc validates that the user contained in
// the headers can operate in the organization. In the case of a jwt it's
// validated with the key returned by keyfunc.
func ValidateOrganizationHeadersWithKeyfunc(ctx context.Context, db *sql.DB, organizationName string, headers http.Header, keyfunc jwt.Keyfunc) (organizationId, userId int64, profileName string, err error) {
	apiKey, err := GetApiKeyWithKeyfunc(headers, keyfunc)

	if err != nil {
		return 0, 0, "", err
	}
	return ValidateOrganizationApiKey(ctx, db, organizationName, apiKey)
}
