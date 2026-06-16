package app

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"stock-management/backend/config"
)

var JwtKey = []byte(config.C.JWTSecret)

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func AllowOriginMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	}
}

func MidDecodeJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			JSONErrorUnauthorized(c, errors.New("missing authorization header"))
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			JSONErrorUnauthorized(c, errors.New("invalid authorization format"))
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return JwtKey, nil
		})

		if err != nil || !token.Valid {
			JSONErrorUnauthorized(c, errors.New("invalid or expired token"))
			c.Abort()
			return
		}

		session := &UserSession{
			Username: claims.Username,
			Role:     claims.Role,
		}

		c.Set(string(UserContextKey), session)
		c.Next()
	}
}

func MidRequireRole(allowedRoles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionVal, exists := c.Get(string(UserContextKey))
		if !exists {
			JSONErrorUnauthorized(c, errors.New("unauthorized user context"))
			c.Abort()
			return
		}

		session, ok := sessionVal.(*UserSession)
		if !ok {
			JSONErrorUnauthorized(c, errors.New("invalid user context payload"))
			c.Abort()
			return
		}

		isAllowed := false
		for _, r := range allowedRoles {
			if session.Role == r {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			JSONErrorForbidden(c, errors.New("insufficient privileges for role "+session.Role))
			c.Abort()
			return
		}

		c.Next()
	}
}
