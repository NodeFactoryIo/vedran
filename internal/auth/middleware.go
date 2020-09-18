package auth

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwtToken := r.Header.Get("X-Auth-Token")

		token, err := jwt.ParseWithClaims(jwtToken, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(authSecret), nil
		})

		if token == nil {
			// todo
			return
		}

		if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
			ctx := context.WithValue(r.Context(), "node-id", claims.NodeId)
			// Access context values in handlers like this
			// props, _ := r.Context().Value("props").(jwt.MapClaims)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			fmt.Println(err)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
		}
	})
}