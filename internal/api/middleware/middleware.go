package middleware

import (
	"log"
	"net/http"
	"pet-project/pkg/config"
	"pet-project/pkg/myJwt"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
)

func RoleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "Нет токена", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.ParseWithClaims(tokenStr, &myJwt.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return config.SecretKey, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Неверный токен", http.StatusUnauthorized)
			log.Printf("Неверный токен: %v,%v", token, err)
			return
		}
		role := token.Claims.(*myJwt.CustomClaims).Role
		switch r.Method {
		case "POST", "PUT", "DELETE":
			if role != "admin" {
				http.Error(w, "Недостаточно прав", http.StatusForbidden)
				return
			}
		default:
		}
		next.ServeHTTP(w, r)
	})
}
