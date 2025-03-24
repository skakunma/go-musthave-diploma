package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/skakunma/go-musthave-diploma-tpl/internal/config"
	jwtauth "github.com/skakunma/go-musthave-diploma-tpl/internal/jwt"
)

var authPaths = map[string]struct{}{
	"/api/user/orders":           {},
	"/api/user/profile":          {},
	"/api/user/balance/withdraw": {},
}

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем, защищен ли маршрут
		if _, exist := authPaths[c.Request.URL.Path]; exist {
			var jwtToken string
			var err error

			// Пробуем взять токен из cookie
			jwtToken, err = c.Cookie("jwt")

			// Если в cookie нет токена, пробуем из заголовка Authorization
			if err != nil || jwtToken == "" {
				authHeader := c.GetHeader("Authorization")
				if authHeader != "" {
					parts := strings.Split(authHeader, " ")
					if len(parts) == 2 && (parts[0] == "Bearer" || parts[0] == "Token") {
						jwtToken = parts[1]
					}
				}
			}

			// Если токен так и не нашли — отправляем 401
			if jwtToken == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "You are not authorized"})
				return
			}

			// Разбираем токен
			claims := &jwtauth.Claims{}
			token, err := jwt.ParseWithClaims(jwtToken, claims, func(t *jwt.Token) (interface{}, error) {
				return []byte(jwtauth.SecretKEY), nil
			})

			// Проверяем валидность токена
			if err != nil || !token.Valid {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Недействительный токен"})
				return
			}

			// Сохраняем информацию о пользователе в контекст
			c.Set("user", claims)
		}

		c.Next()
	}
}
