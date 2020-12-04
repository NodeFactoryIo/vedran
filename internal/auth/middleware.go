package auth

import (
	"context"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type ContextKey string

const RequestContextKey = ContextKey("request")

type RequestContext struct {
	NodeId    string
	Timestamp time.Time
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwtToken := r.Header.Get("X-Auth-Header")
		token, err := ParseJwtTokenWithCustomClaims(jwtToken)
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

		log.Errorf("Unauthorized request: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Unauthorized"))
	})
}
