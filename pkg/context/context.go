package context

import (
"context"
"net/http"
)

type key string

const UserIDKey key = "user_id"

// GetUserID извлекает user_id из контекста запроса
func GetUserID(r *http.Request) int {
if userID, ok := r.Context().Value(UserIDKey).(int); ok {
return userID
}
return 0
}

// SetUserID добавляет user_id в контекст запроса
func SetUserID(r *http.Request, userID int) *http.Request {
ctx := context.WithValue(r.Context(), UserIDKey, userID)
return r.WithContext(ctx)
}
