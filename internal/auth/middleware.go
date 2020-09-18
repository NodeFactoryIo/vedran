package auth

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"time"
)

type ContextKey string

const RequestContextKey = ContextKey("request")

type RequestContext struct {
	NodeId string
	Timestamp time.Time
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwtToken := r.Header.Get("X-Auth-Token")

		token, err := jwt.ParseWithClaims(jwtToken, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(authSecret), nil
		})

		if err == nil {
			if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
				c := &RequestContext{
					NodeId:    claims.NodeId,
					Timestamp: time.Now(),
				}
				ctx := context.WithValue(r.Context(), RequestContextKey, c)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		log.Println("Unauthorized request:", err)
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Unauthorized"))
	})
}
