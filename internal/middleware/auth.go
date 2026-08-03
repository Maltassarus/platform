package middleware

import (
"net/http"
"strings"

"platform/pkg/auth"
"platform/pkg/context"
"platform/pkg/response"
)

type AuthMiddleware struct {
jwtManager *auth.JWTManager
}

func NewAuthMiddleware(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
return func(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
authHeader := r.Header.Get("Authorization")
if authHeader == "" {
response.Error(w, http.StatusUnauthorized, "Authorization header required")
return
}

parts := strings.Split(authHeader, " ")
if len(parts) != 2 || parts[0] != "Bearer" {
response.Error(w, http.StatusUnauthorized, "Invalid authorization header format")
return
}

userID, err := jwtManager.ValidateToken(parts[1])
if err != nil {
response.Error(w, http.StatusUnauthorized, "Invalid or expired token")
return
}

r = context.SetUserID(r, userID)
next.ServeHTTP(w, r)
})
}
}
