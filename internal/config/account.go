package config

import "strings"

const BrowserSessionAuthMode = "browser_session"

func (a Account) Identifier() string {
	if strings.TrimSpace(a.Email) != "" {
		return strings.TrimSpace(a.Email)
	}
	if mobile := NormalizeMobileForStorage(a.Mobile); mobile != "" {
		return mobile
	}
	if a.IsBrowserSession() {
		if profile := strings.TrimSpace(a.ProfileID); profile != "" {
			return "profile:" + profile
		}
	}
	return ""
}

func (a Account) IsBrowserSession() bool {
	return strings.EqualFold(strings.TrimSpace(a.AuthMode), BrowserSessionAuthMode) && strings.TrimSpace(a.ProfileID) != ""
}
