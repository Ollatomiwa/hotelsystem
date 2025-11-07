package security

import (
	"errors"
	"fmt"
	"time"
	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secretKey string
	refreshKey string
	accessTokenDuration time.Duration
	refreshTokenDuration time.Duration
}

type Claims struct  {
	UserId string `json:"userId"`
	Role string `json:"role"`
	jwt.RegisteredClaims 
}

type RefreshClaims struct {
	UserId string `json:"userId"`
	jwt.RegisteredClaims 
}

func NewJWTManager(secretKey, refreshKey string, accessDuration, refreshDuration time.Duration) *JWTManager{
	return &JWTManager{
		secretKey: secretKey,
		refreshKey: refreshKey,
		accessTokenDuration: accessDuration,
		refreshTokenDuration: refreshDuration,
	}
}

func (m *JWTManager) GenerateAccessToken(userId, role string) (string, error) {
	claims := &Claims{
		UserId : userId,
		Role : role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessTokenDuration)),
			IssuedAt : jwt.NewNumericDate(time.Now()),
			Issuer : "user-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

func (m *JWTManager) GenerateRefreshToken(userId string) (string, error) {
	claims := &RefreshClaims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.refreshTokenDuration)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
			Issuer: "user-service",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.refreshKey))
}

func (m *JWTManager) VerifyAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token)(interface{}, error){
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.secretKey), nil 
	})

	if err != nil {
		return nil, err 
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil 
	}
	return nil, errors.New("invalid token")
}

func (m *JWTManager) VerifyRefreshToken(tokenString string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token)(interface{}, error){
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.refreshKey), nil 
	})

	if err != nil {
		return nil, err 
	}

	if claims, ok := token.Claims.(*RefreshClaims); ok && token.Valid {
		return claims, nil 
	}
	return nil, errors.New("invalid refresh token")
}

func (m *JWTManager) RefreshTokens(refreshToken string) (string, string, error) {
	claims, err := m.VerifyRefreshToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	newAccessToken, err := m.GenerateAccessToken(claims.UserId, "")
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w",err)
	}
	newRefreshToken, err := m.GenerateRefreshToken(claims.UserId)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return newAccessToken, newRefreshToken, nil 
}

