package server

import (
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const APIKeyHeader = "api_key"

// LoggedHandler logs interesting request and response attributes whenever a request is
// received or completed.
func LoggedHandler(s *slog.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		attrs := []slog.Attr{
			slog.String("method", r.Method),
			slog.String("host", r.Host),
			slog.String("path", r.URL.Path),
			slog.String("query", r.URL.RawQuery),
			slog.String("ip", r.RemoteAddr),
			slog.String("referer", r.Referer()),
			slog.String("user-agent", r.UserAgent()),
		}

		s.LogAttrs(r.Context(), slog.LevelInfo, "request start", attrs...)

		h.ServeHTTP(w, r)

		defer func() {
			end := time.Now()
			attrs = append(attrs, slog.Duration("latency", end.Sub(start)))
			s.LogAttrs(r.Context(), slog.LevelInfo, "request complete", attrs...)
		}()
	})
}

// AuthenticatedHandler checks the incoming request for an authentication token rejecting
// the request if not found or not valid. The auth token is then propagated for future use.
//
// except is a basic path prefix that lists routes that should be public and not require authentication
func AuthenticatedHandler(p StaticAuthProvider, s *slog.Logger, next http.Handler, excepts ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, except := range excepts {
			// skip auth if the route is excepted
			if strings.HasPrefix(r.URL.Path, except) {
				next.ServeHTTP(w, r)
				return
			}
		}

		token, has := p[r.Header.Get(APIKeyHeader)]
		if !has {
			s.Log(r.Context(), slog.LevelWarn, "request does not contain expected credentions")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if err := token.Validate(time.Now()); err != nil {
			s.Log(r.Context(), slog.LevelWarn, "invalid token presented: "+err.Error())
			w.WriteHeader(http.StatusForbidden)
			return
		}
		r = r.WithContext(token.Ctx(r.Context()))
		next.ServeHTTP(w, r)
	})
}

// ScopedHandler looks for scope within a [Token] found in [r.Context()] rejecting the request
// if not found.
func ScopedHandler(s *slog.Logger, scope string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, has := TokenFromContext(r.Context())
		if !has {
			s.Log(r.Context(), slog.LevelError, "scoped handler didn't find a token, router set up incorrectly")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if _, has = token.Scopes[scope]; !has {
			s.Log(r.Context(), slog.LevelWarn, "token missing required scope: "+scope)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
