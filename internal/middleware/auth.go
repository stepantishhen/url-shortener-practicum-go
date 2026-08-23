package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

type contextKey string

const (
	UserIDKey        contextKey = "userID"
	CookieInvalidKey contextKey = "cookieInvalid"
	cookieName                  = "user_id"
)

func Auth(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			cookie, err := r.Cookie(cookieName)
			if err != nil {
				// Cookie absent — issue a new one.
				userID, err := newUserID()
				if err != nil {
					http.Error(w, "internal error", http.StatusInternalServerError)
					return
				}
				setAuthCookie(w, userID, secretKey)
				ctx = context.WithValue(ctx, UserIDKey, userID)
				ctx = context.WithValue(ctx, CookieInvalidKey, false)
			} else {
				userID, valid := parseSignedCookie(cookie.Value, secretKey)
				if valid {
					ctx = context.WithValue(ctx, UserIDKey, userID)
					ctx = context.WithValue(ctx, CookieInvalidKey, false)
				} else {
					// Cookie present but signature invalid — issue a new one.
					userID, err := newUserID()
					if err != nil {
						http.Error(w, "internal error", http.StatusInternalServerError)
						return
					}
					setAuthCookie(w, userID, secretKey)
					ctx = context.WithValue(ctx, UserIDKey, userID)
					ctx = context.WithValue(ctx, CookieInvalidKey, true)
				}
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func parseSignedCookie(value, secretKey string) (string, bool) {
	parts := strings.SplitN(value, "|", 2)
	if len(parts) != 2 || parts[0] == "" {
		return "", false
	}
	userID, sigB64 := parts[0], parts[1]
	sigBytes, err := base64.URLEncoding.DecodeString(sigB64)
	if err != nil {
		return "", false
	}
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(userID))
	expected := mac.Sum(nil)
	if !hmac.Equal(sigBytes, expected) {
		return "", false
	}
	return userID, true
}

func setAuthCookie(w http.ResponseWriter, userID, secretKey string) {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(userID))
	sig := base64.URLEncoding.EncodeToString(mac.Sum(nil))
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    userID + "|" + sig,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func newUserID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

// UserID extracts the user ID from ctx. Returns empty string if not present.
func UserID(ctx context.Context) string {
	id, _ := ctx.Value(UserIDKey).(string)
	return id
}
