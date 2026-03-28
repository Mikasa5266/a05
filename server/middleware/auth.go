package middleware

import (
	"net/http"
	"strings"
	"time"

	"your-project/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(config.GetConfig().JWT.Secret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userIDRaw, ok := claims["user_id"]
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token payload"})
				c.Abort()
				return
			}

			userIDFloat, ok := userIDRaw.(float64)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token payload"})
				c.Abort()
				return
			}

			roleRaw, ok := claims["role"]
			role, roleOk := roleRaw.(string)
			if !ok || !roleOk || role == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token payload"})
				c.Abort()
				return
			}

			c.Set("user_id", uint(userIDFloat))
			c.Set("role", role)
			if userUUID, ok := claims["user_uuid"].(string); ok && strings.TrimSpace(userUUID) != "" {
				c.Set("user_uuid", strings.TrimSpace(userUUID))
			}
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
		}
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		normalized := strings.TrimSpace(strings.ToLower(role))
		if normalized != "" {
			allowed[normalized] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		role := strings.TrimSpace(strings.ToLower(c.GetString("role")))
		if role == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Role required"})
			c.Abort()
			return
		}

		if _, ok := allowed[role]; !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient role permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func GenerateToken(userID uint, role, userUUID string) (string, error) {
	config := config.GetConfig()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   userID,
		"role":      role,
		"user_uuid": strings.TrimSpace(userUUID),
		"exp":       time.Now().Add(time.Hour * time.Duration(config.JWT.ExpireTime)).Unix(),
	})

	return token.SignedString([]byte(config.JWT.Secret))
}
