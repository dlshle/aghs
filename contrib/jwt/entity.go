package jwt

import (
	"time"

	"github.com/golang-jwt/jwt"
)

const (
	ClaimKeyIssuer    = "Issuer"
	ClaimKeyExpiresAt = "ExpiresAt"
)

type TokenClaim map[string]interface{}

func toTokenClaim(mapClaim jwt.MapClaims) TokenClaim {
	return TokenClaim(mapClaim)
}

func (c TokenClaim) IsExpired() bool {
	if expiresAt, exists := c[ClaimKeyExpiresAt]; exists {
		return time.Unix(expiresAt.(int64), 0).Before(time.Now())
	}
	return false
}

func (c TokenClaim) IsIssuedByMe(myIssueKey string) bool {
	if issuer, exists := c[ClaimKeyIssuer]; exists {
		return issuer.(string) == myIssueKey
	}
	return false
}
