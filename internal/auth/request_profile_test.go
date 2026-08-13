package auth

import (
	"context"
	"testing"

	"ds2api/internal/config"
)

func TestEnsureManagedTokenSkipsBrowserSessionLogin(t *testing.T) {
	called := false
	r := &Resolver{Login: func(context.Context, config.Account) (string, error) {
		called = true
		return "", nil
	}}
	a := &RequestAuth{
		AccountID: "profile:p1",
		Account: config.Account{
			AuthMode:  config.BrowserSessionAuthMode,
			ProfileID: "p1",
		},
	}
	if err := r.ensureManagedToken(context.Background(), a); err != nil {
		t.Fatalf("ensureManagedToken() error = %v", err)
	}
	if called {
		t.Fatal("password login should not run for browser-session accounts")
	}
	if a.DeepSeekToken != "" {
		t.Fatalf("DeepSeekToken = %q", a.DeepSeekToken)
	}
}
