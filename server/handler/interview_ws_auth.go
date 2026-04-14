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

func normalizeLiveRoleAlias(role string) string {
	normalized := strings.TrimSpace(strings.ToLower(role))
	switch normalized {
	case "teacher", "mentor", "faculty":
		return "university"
	case "hr", "interviewer", "recruiter":
		return "enterprise"
	default:
		return normalized
	}
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
	normalizedRole := normalizeLiveRoleAlias(role)
	if !roleOK || normalizedRole == "" {
		return nil, jwt.ErrTokenInvalidClaims
	}

	userUUID, _ := claims["user_uuid"].(string)

	return &liveTokenIdentity{
		UserID:   uint(uid),
		Role:     normalizedRole,
		UserUUID: strings.TrimSpace(userUUID),
	}, nil
}
