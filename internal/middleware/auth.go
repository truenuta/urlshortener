package middleware

import (
	"context"
	"net/http"

	"github.com/truenuta/urlshortener/internal/auth"
)

type ctxKey string

const (
	userIDKey     ctxKey = "id"
	authFailedKey ctxKey = "authFailed"
)

func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok
}

func AuthFailedFromContext(ctx context.Context) bool {
	failed, _ := ctx.Value(authFailedKey).(bool)
	return failed
}

func AuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var id string

		authFailed := false
		cookies, getCookiesErr := r.Cookie("JWTtoken")

		if getCookiesErr != nil {
			newID, err := auth.GenerateUserID()
			if err == nil {
				token, tokenErr := auth.BuildJWTString(newID)
				id = newID
				if tokenErr == nil {
					http.SetCookie(w, &http.Cookie{Name: "JWTtoken", Value: token})
				}
			}
		} else {
			parsedID, parsedErr := auth.GetUserID(cookies.Value)
			if parsedErr == nil && parsedID != "" {
				id = parsedID
			} else {
				authFailed = true
			}
		}
		ctx := context.WithValue(r.Context(), userIDKey, id)
		ctx = context.WithValue(ctx, authFailedKey, authFailed)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)

	})
}
