package resolvers

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/httpx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

const (
	// refreshCookieName carries the raw refresh token for browser clients. It is
	// httpOnly so it is never exposed to JavaScript.
	refreshCookieName = "refresh_token"
	// csrfCookieName carries the CSRF token. It is readable by JavaScript so the
	// SPA can echo it in the X-CSRF-Token header (the double-submit pattern).
	csrfCookieName = "csrf_token"
	// csrfHeaderName is the header the client must echo the CSRF cookie value in.
	csrfHeaderName = "X-CSRF-Token"
	// cookiePath is the root path so the JS-readable CSRF cookie is visible to the SPA on
	// every route (document.cookie is scoped by path), not only under the GraphQL endpoint.
	cookiePath = "/"
	// csrfTokenBytes is the entropy of a generated CSRF token before encoding.
	csrfTokenBytes = 32
)

// setAuthCookies issues the httpOnly refresh cookie and the JS-readable CSRF
// cookie. It is a no-op when no ResponseWriter is available (non-HTTP
// transports such as the websocket), so non-browser flows still work via the
// token returned in the payload body.
func (r *Resolver) setAuthCookies(ctx context.Context, refreshToken string) error {
	w, ok := httpx.ResponseWriter(ctx)
	if !ok {
		return nil
	}

	csrfToken, err := randomToken()
	if err != nil {
		return err
	}

	maxAge := int(r.RefreshTTL.Seconds())

	// Secure is environment-driven (COOKIE_SECURE=true in production); SameSite=Strict
	// and HttpOnly guard the session cookie regardless.
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    refreshToken,
		Path:     cookiePath,
		Domain:   r.CookieDomain,
		MaxAge:   maxAge,
		Secure:   r.CookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	// The CSRF cookie is intentionally readable by JavaScript so the client can echo it
	// back as a header for the double-submit check.
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    csrfToken,
		Path:     cookiePath,
		Domain:   r.CookieDomain,
		MaxAge:   maxAge,
		Secure:   r.CookieSecure,
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
	})
	return nil
}

// clearAuthCookies expires both auth cookies.
func (r *Resolver) clearAuthCookies(ctx context.Context) {
	w, ok := httpx.ResponseWriter(ctx)
	if !ok {
		return
	}
	for _, name := range []string{refreshCookieName, csrfCookieName} {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     cookiePath,
			Domain:   r.CookieDomain,
			MaxAge:   -1,
			Secure:   r.CookieSecure,
			HttpOnly: name == refreshCookieName,
			SameSite: http.SameSiteStrictMode,
		})
	}
}

// readRefreshCookie returns the raw refresh token from the request cookie, or
// the empty string when absent.
func readRefreshCookie(ctx context.Context) string {
	req, ok := httpx.Request(ctx)
	if !ok {
		return ""
	}
	c, err := req.Cookie(refreshCookieName)
	if err != nil {
		return ""
	}
	return c.Value
}

// verifyCSRF enforces the double-submit cookie check: the X-CSRF-Token header
// must be present and equal (in constant time) to the csrf_token cookie.
func verifyCSRF(ctx context.Context) error {
	req, ok := httpx.Request(ctx)
	if !ok {
		return errors.Unauthenticated("CSRF_INVALID", "csrf verification unavailable")
	}
	header := req.Header.Get(csrfHeaderName)
	cookie, err := req.Cookie(csrfCookieName)
	if err != nil || header == "" || cookie.Value == "" {
		return errors.Unauthenticated("CSRF_INVALID", "missing csrf token")
	}
	if subtle.ConstantTimeCompare([]byte(header), []byte(cookie.Value)) != 1 {
		return errors.Unauthenticated("CSRF_INVALID", "csrf token mismatch")
	}
	return nil
}

// randomToken returns a URL-safe, base64-encoded random token.
func randomToken() (string, error) {
	buf := make([]byte, csrfTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", errors.Wrap(err, errors.KindInternal, "CSRF_GENERATE_FAILED", "failed to generate csrf token")
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
