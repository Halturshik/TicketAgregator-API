package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func (j *JWTService) GenerateAccessToken(userID int, version int) (string, error) {
	return j.generateToken(userID, version, TypeAccess, AccessTokenTTL)
}

func (j *JWTService) GenerateRefreshToken(userID int, version int) (string, error) {
	return j.generateToken(userID, version, TypeRefresh, RefreshTokenTTL)
}

func (j *JWTService) generateToken(userID int, version int, tokenType string, ttl time.Duration) (string, error) {
	claims := &Claims{
		UserID:  userID,
		Version: version,
		Type:    tokenType,
		JTI:     uuid.NewString(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTService) ParseAccessToken(tokenStr string) (int, int, error) {
	return j.parseToken(tokenStr, TypeAccess)
}

func (j *JWTService) ParseRefreshToken(tokenStr string) (int, int, error) {
	return j.parseToken(tokenStr, TypeRefresh)
}

func (j *JWTService) parseToken(tokenStr string, expectedType string) (int, int, error) {
	parsed, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Method.Alg())
		}
		return j.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return 0, 0, err
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return 0, 0, errors.New("invalid token")
	}
	if claims.Type != expectedType {
		return 0, 0, errors.New("invalid token type")
	}

	return claims.UserID, claims.Version, nil
}
