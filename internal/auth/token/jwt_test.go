package token

import (
	"testing"

	"github.com/golang-jwt/jwt/v4"
)

func TestAccessAndRefreshTokensCannotBeInterchanged(t *testing.T) {
	service := NewJWTService("test-secret")
	accessToken, err := service.GenerateAccessToken(42, 7)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	refreshToken, err := service.GenerateRefreshToken(42, 7)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	assertParsedToken(t, service.ParseAccessToken, accessToken, 42, 7)
	assertParsedToken(t, service.ParseRefreshToken, refreshToken, 42, 7)
	if _, _, err := service.ParseAccessToken(refreshToken); err == nil {
		t.Fatal("refresh token was accepted as access token")
	}
	if _, _, err := service.ParseRefreshToken(accessToken); err == nil {
		t.Fatal("access token was accepted as refresh token")
	}
}

func TestJWTRejectsUnexpectedSigningMethodAndSecret(t *testing.T) {
	service := NewJWTService("test-secret")
	claims := Claims{UserID: 42, Version: 1, Type: TypeAccess}
	wrongMethod, err := jwt.NewWithClaims(jwt.SigningMethodHS384, claims).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign HS384 token: %v", err)
	}
	if _, _, err := service.ParseAccessToken(wrongMethod); err == nil {
		t.Fatal("HS384 token was accepted by HS256-only parser")
	}

	otherService := NewJWTService("other-secret")
	accessToken, err := otherService.GenerateAccessToken(42, 1)
	if err != nil {
		t.Fatalf("generate token with another secret: %v", err)
	}
	if _, _, err := service.ParseAccessToken(accessToken); err == nil {
		t.Fatal("token signed with another secret was accepted")
	}
}

func TestGeneratedTokensAreUnique(t *testing.T) {
	service := NewJWTService("test-secret")
	first, err := service.GenerateAccessToken(42, 1)
	if err != nil {
		t.Fatalf("generate first token: %v", err)
	}
	second, err := service.GenerateAccessToken(42, 1)
	if err != nil {
		t.Fatalf("generate second token: %v", err)
	}
	if first == second {
		t.Fatal("sequential tokens are equal; JTI uniqueness is broken")
	}
}

func assertParsedToken(
	t *testing.T,
	parse func(string) (int, int, error),
	token string,
	wantUserID int,
	wantVersion int,
) {
	t.Helper()
	userID, version, err := parse(token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if userID != wantUserID || version != wantVersion {
		t.Fatalf("parsed token = userID:%d version:%d, want userID:%d version:%d", userID, version, wantUserID, wantVersion)
	}
}
