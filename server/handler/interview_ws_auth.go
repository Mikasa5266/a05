package handler

import (
	"strings"

	"your-project/config"

	"github.com/golang-jwt/jwt/v5"
)

type liveTokenIdentity struct {
	UserID   uint
	Role     string
	UserUUID string
}

func parseLiveIdentityFromToken(tokenString string) (*liveTokenIdentity, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(config.GetConfig().JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	uid, ok := claims["user_id"].(float64)
	if !ok || uid <= 0 {
		return nil, jwt.ErrTokenInvalidClaims
	}

	role, roleOK := claims["role"].(string)
	if !roleOK || strings.TrimSpace(role) == "" {
		return nil, jwt.ErrTokenInvalidClaims
	}

	userUUID, _ := claims["user_uuid"].(string)

	return &liveTokenIdentity{
		UserID:   uint(uid),
		Role:     strings.TrimSpace(role),
		UserUUID: strings.TrimSpace(userUUID),
	}, nil
}
