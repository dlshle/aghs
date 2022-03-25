package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

func createTokenClaim(issuer string, rawClaims map[string]interface{}, ttl time.Duration) jwt.MapClaims {
	claims := make(map[string]interface{})
	claims[ClaimKeyIssuer] = issuer
	claims[ClaimKeyExpiresAt] = time.Now().Add(ttl).Unix()
	for k, v := range rawClaims {
		claims[k] = v
	}
	return claims
}

func SignToken(issuer string, claims map[string]interface{}, secrect string, ttl time.Duration) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, createTokenClaim(issuer, claims, ttl)).
		SignedString([]byte(secrect))
}

type TokenVerificationFunc func(claims TokenClaim) (string, error) // returns secret and error

func VerifyToken(stringToken string, verifyCallback TokenVerificationFunc) (*jwt.Token, error) {
	return jwt.ParseWithClaims(stringToken, jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		tokenClaims := toTokenClaim(t.Claims.(jwt.MapClaims))
		secret, err := verifyCallback(tokenClaims)
		if err != nil {
			return nil, err
		}
		return []byte(secret), nil
	})
}
