package auth

import (
	"testing"

	"github.com/PietroTelino/go-api-template/internal/config"
)

func newTestTokenManager(accessTokenSeconds int) *TokenManager {
	return NewTokenManager(&config.Config{
		JWTSecret:                  "segredo-de-teste-access",
		JWTExpiresInSeconds:        accessTokenSeconds,
		JWTRefreshSecret:           "segredo-de-teste-refresh",
		JWTRefreshExpiresInSeconds: 3600,
	})
}

func TestAccessToken(t *testing.T) {
	tokenManager := newTestTokenManager(900)

	accessToken, err := tokenManager.GenerateAccessToken("id-do-usuario", "administrator")
	if err != nil {
		t.Fatalf("erro ao gerar o access token: %v", err)
	}

	claims, err := tokenManager.ParseAccessToken(accessToken)
	if err != nil {
		t.Fatalf("erro ao ler o access token: %v", err)
	}

	if claims.Subject != "id-do-usuario" || claims.Role != "administrator" {
		t.Errorf("claims = (%q, %q), esperado (%q, %q)", claims.Subject, claims.Role, "id-do-usuario", "administrator")
	}
}

func TestExpiredAccessToken(t *testing.T) {
	tokenManager := newTestTokenManager(-60)

	accessToken, err := tokenManager.GenerateAccessToken("id-do-usuario", "user")
	if err != nil {
		t.Fatalf("erro ao gerar o access token: %v", err)
	}

	if _, err := tokenManager.ParseAccessToken(accessToken); err == nil {
		t.Error("um access token expirado foi aceito")
	}
}

func TestTokenSignedWithAnotherSecret(t *testing.T) {
	otherTokenManager := NewTokenManager(&config.Config{JWTSecret: "outro-segredo", JWTExpiresInSeconds: 900})

	accessToken, err := otherTokenManager.GenerateAccessToken("id-do-usuario", "god")
	if err != nil {
		t.Fatalf("erro ao gerar o access token: %v", err)
	}

	if _, err := newTestTokenManager(900).ParseAccessToken(accessToken); err == nil {
		t.Error("um token assinado com outro segredo foi aceito")
	}
}

func TestAccessAndRefreshTokensAreNotInterchangeable(t *testing.T) {
	tokenManager := newTestTokenManager(900)

	accessToken, _ := tokenManager.GenerateAccessToken("id-do-usuario", "user")
	refreshToken, _ := tokenManager.GenerateRefreshToken("id-do-usuario")

	if _, err := tokenManager.ParseRefreshToken(accessToken); err == nil {
		t.Error("um access token foi aceito como refresh token")
	}

	if _, err := tokenManager.ParseAccessToken(refreshToken); err == nil {
		t.Error("um refresh token foi aceito como access token")
	}
}

func TestRefreshTokensAreUnique(t *testing.T) {
	tokenManager := newTestTokenManager(900)

	firstToken, _ := tokenManager.GenerateRefreshToken("id-do-usuario")
	secondToken, _ := tokenManager.GenerateRefreshToken("id-do-usuario")

	if firstToken == secondToken {
		t.Error("dois refresh tokens gerados no mesmo segundo ficaram iguais")
	}

	claims, err := tokenManager.ParseRefreshToken(firstToken)
	if err != nil {
		t.Fatalf("erro ao ler o refresh token: %v", err)
	}

	if claims.Subject != "id-do-usuario" || claims.ID == "" {
		t.Errorf("claims inesperadas: subject=%q id=%q", claims.Subject, claims.ID)
	}
}
