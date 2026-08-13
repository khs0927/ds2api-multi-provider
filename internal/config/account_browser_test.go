package config

import "testing"

func TestBrowserSessionAccountIdentifier(t *testing.T) {
	acc := Account{AuthMode: BrowserSessionAuthMode, ProfileID: "deepseek-google-01"}
	if got := acc.Identifier(); got != "profile:deepseek-google-01" {
		t.Fatalf("Identifier() = %q", got)
	}
	if !acc.IsBrowserSession() {
		t.Fatal("expected browser-session account")
	}
}

func TestLegacyIdentifierStillWins(t *testing.T) {
	acc := Account{Email: "user@example.com", AuthMode: BrowserSessionAuthMode, ProfileID: "p1"}
	if got := acc.Identifier(); got != "user@example.com" {
		t.Fatalf("Identifier() = %q", got)
	}
}

func TestDropInvalidAccountsKeepsProfileAccount(t *testing.T) {
	cfg := Config{Accounts: []Account{
		{AuthMode: BrowserSessionAuthMode, ProfileID: "p1"},
		{AuthMode: BrowserSessionAuthMode},
	}}
	cfg.DropInvalidAccounts()
	if len(cfg.Accounts) != 1 || cfg.Accounts[0].ProfileID != "p1" {
		t.Fatalf("accounts = %#v", cfg.Accounts)
	}
}
