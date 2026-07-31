package auth

import (
	"fmt"
	"time"

	"github.com/PietroTelino/go-api-template/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

type AccessTokenClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	accessTokenSecret  []byte
	accessTokenTTL     time.Duration
	refreshTokenSecret []byte
	refreshTokenTTL    time.Duration
}

func NewTokenManager(appConfig *config.Config) *TokenManager {
	return &TokenManager{
		accessTokenSecret:  []byte(appConfig.JWTSecret),
		accessTokenTTL:     time.Duration(appConfig.JWTExpiresInSeconds) * time.Second,
		refreshTokenSecret: []byte(appConfig.JWTRefreshSecret),
		refreshTokenTTL:    time.Duration(appConfig.JWTRefreshExpiresInSeconds) * time.Second,
	}
}

func (tokenManager *TokenManager) RefreshTokenTTL() time.Duration {
	return tokenManager.refreshTokenTTL
}

func (tokenManager *TokenManager) GenerateAccessToken(userID string, userRole string) (string, error) {
	claims := AccessTokenClaims{
		Role: userRole,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenManager.accessTokenTTL)),
		},
	}

	unsignedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return unsignedToken.SignedString(tokenManager.accessTokenSecret)
}

func (tokenManager *TokenManager) GenerateRefreshToken(userID string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenManager.refreshTokenTTL)),
	}

	unsignedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return unsignedToken.SignedString(tokenManager.refreshTokenSecret)
}

func (tokenManager *TokenManager) ParseAccessToken(accessTokenString string) (*AccessTokenClaims, error) {
	claims := &AccessTokenClaims{}

	parsedToken, err := jwt.ParseWithClaims(accessTokenString, claims, func(token *jwt.Token) (any, error) {
		if _, signingMethodIsHMAC := token.Method.(*jwt.SigningMethodHMAC); !signingMethodIsHMAC {
			return nil, fmt.Errorf("Método de assinatura inesperado: %v", token.Header["alg"])
		}
		return tokenManager.accessTokenSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !parsedToken.Valid {
		return nil, fmt.Errorf("Access token inválido")
	}

	return claims, nil
}

func (tokenManager *TokenManager) ParseRefreshToken(refreshTokenString string) (*jwt.RegisteredClaims, error) {
	claims := &jwt.RegisteredClaims{}

	parsedToken, err := jwt.ParseWithClaims(refreshTokenString, claims, func(token *jwt.Token) (any, error) {
		if _, signingMethodIsHMAC := token.Method.(*jwt.SigningMethodHMAC); !signingMethodIsHMAC {
			return nil, fmt.Errorf("Método de assinatura inesperado: %v", token.Header["alg"])
		}
		return tokenManager.refreshTokenSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !parsedToken.Valid {
		return nil, fmt.Errorf("Refresh token inválido")
	}

	return claims, nil
}
