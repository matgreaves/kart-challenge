package server

import (
	"context"
	"errors"
	"time"
)

type tokenKey struct{}

// Token is a really basic example of a session auth token we might use within our application.
//
// There are lots of alternatives with the defacto standard being the [JWT](https://datatracker.ietf.org/doc/html/rfc7519).
type Token struct {
	ValidFrom time.Time
	ExpiresAt time.Time
	Scopes    map[string]struct{}
}

// Ctx stores t in ctx.
func (t Token) Ctx(ctx context.Context) context.Context {
	return context.WithValue(ctx, tokenKey{}, t)
}

// TokenFromContext returns a [Token] if present in ctx in which case found will be true.
func TokenFromContext(ctx context.Context) (_ Token, found bool) {
	v := ctx.Value(tokenKey{})
	if v == nil {
		return Token{}, false
	}
	return v.(Token), true
}

// Validate checks whether the token should be accepted for further use.
func (t Token) Validate(at time.Time) error {
	if at.Before(t.ValidFrom) {
		return errors.New("token presented before ValidFrom")
	}
	if at.After(t.ExpiresAt) {
		return errors.New("token presented after ExpiresAt")
	}
	return nil
}

// StaticAuthProvider is an obviously very insecure way to manage tokens used as
// an example of how we can hook authentication into our workflow. It goes without saying
// this shouldn't ever be included in any code that gets deployed to a live environment.
//
// We're not providing an abstraction over this provider as interaction patters vary wildly depending on
// the authentication scheme used.
type StaticAuthProvider map[string]Token

// TestAuth has a few prebaked api tokens useful for checking authentication handlin, needless to say this should
// never be used in a deployed application.
func TestAuth() StaticAuthProvider {
	return StaticAuthProvider{
		"apitest":  Token{ExpiresAt: time.Now().AddDate(1, 0, 0), Scopes: map[string]struct{}{"order:create": {}}},
		"noscope":  Token{ExpiresAt: time.Now().AddDate(1, 0, 0)},
		"tooearly": Token{ValidFrom: time.Now().AddDate(1, 0, 0), ExpiresAt: time.Now().AddDate(1, 0, 0)},
		"toolate":  Token{ValidFrom: time.Now().AddDate(-1, 0, 0), ExpiresAt: time.Now().AddDate(-1, 0, 0)},
	}
}
