package token

import "github.com/golang-jwt/jwt/v4"

type JWTService struct {
	secret []byte
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{
		secret: []byte(secret),
	}
}

type Claims struct {
	UserID  int `json:"user_id"`
	Version int `json:"version"`
	jwt.RegisteredClaims
}
