package jwt

import (
	"strings"

	"github.com/vnykmshr/nivo/shared/errors"
)

// ExtractBearerToken extracts the bearer token from the Authorization header.
// Returns the token string or an error if the header is missing or malformed.
func ExtractBearerToken(authHeader string) (string, *errors.Error) {
	if authHeader == "" {
		return "", errors.Unauthorized("missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.Unauthorized("invalid authorization header format")
	}

	return parts[1], nil
}
