package middleware_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"url-shortener-practicum-go/internal/middleware"
)

const testKey = "test-secret-key"

// signedCookie builds a valid cookie value the same way the middleware does.
func signedCookie(userID string) string {
	mac := hmac.New(sha256.New, []byte(testKey))
	mac.Write([]byte(userID))
	return userID + "|" + base64.URLEncoding.EncodeToString(mac.Sum(nil))
}

func applyAuth(r *http.Request) (ctxUserID string, ctxInvalid bool, resp *http.Response) {
	h := middleware.Auth(testKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxUserID, _ = r.Context().Value(middleware.UserIDKey).(string)
		ctxInvalid, _ = r.Context().Value(middleware.CookieInvalidKey).(bool)
	}))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return ctxUserID, ctxInvalid, w.Result()
}

func cookieNamed(resp *http.Response, name string) *http.Cookie {
	for _, c := range resp.Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func TestAuth_NoCookie_IssuesNewCookie(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	userID, invalid, resp := applyAuth(r)

	if userID == "" {
		t.Error("expected non-empty userID in context")
	}
	if invalid {
		t.Error("expected cookieInvalid=false when cookie is absent")
	}
	if c := cookieNamed(resp, "user_id"); c == nil {
		t.Error("expected Set-Cookie: user_id header")
	}
}

func TestAuth_ValidCookie_PreservesUserID(t *testing.T) {
	const existingID = "11111111-1111-4111-8111-111111111111"
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: "user_id", Value: signedCookie(existingID)})

	userID, invalid, resp := applyAuth(r)

	if userID != existingID {
		t.Errorf("expected userID %q, got %q", existingID, userID)
	}
	if invalid {
		t.Error("expected cookieInvalid=false for valid cookie")
	}
	if c := cookieNamed(resp, "user_id"); c != nil {
		t.Error("expected no Set-Cookie for already valid cookie")
	}
}

func TestAuth_InvalidCookie_MarksInvalidAndIssuesNew(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: "user_id", Value: "tampered|invalidsignature=="})

	userID, invalid, resp := applyAuth(r)

	if userID == "" {
		t.Error("expected new userID to be generated")
	}
	if !invalid {
		t.Error("expected cookieInvalid=true for tampered cookie")
	}
	if c := cookieNamed(resp, "user_id"); c == nil {
		t.Error("expected new Set-Cookie: user_id after rejecting invalid cookie")
	}
}

func TestAuth_EmptyUserIDInCookie_MarksInvalid(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: "user_id", Value: "|somesignature"})

	_, invalid, _ := applyAuth(r)

	if !invalid {
		t.Error("expected cookieInvalid=true when cookie has empty user ID part")
	}
}
