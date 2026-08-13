package browserauth

import (
	"context"
	"errors"
	"strings"
	"time"
)

const deepSeekSignInURL = "https://chat.deepseek.com/sign_in"

type Service struct{}

type ProviderStatus struct {
	Configured bool          `json:"configured"`
	BaseURL    string        `json:"base_url,omitempty"`
	TTL        time.Duration `json:"-"`
}

type LoginWindow struct {
	URL string `json:"login_url"`
}

func New() *Service {
	return &Service{}
}

func (s *Service) ProviderStatus() ProviderStatus {
	baseURL, _, ttl, err := providerConfig()
	if err != nil {
		return ProviderStatus{}
	}
	return ProviderStatus{Configured: true, BaseURL: baseURL, TTL: ttl}
}

// OpenLoginWindow attaches to a Browserless profile connection supplied by the
// provider and opens DeepSeek's normal sign-in page. The user completes Google
// sign-in inside that remote browser; DS2API does not collect Google fields.
func (s *Service) OpenLoginWindow(ctx context.Context, connectURL string) (LoginWindow, error) {
	if strings.TrimSpace(connectURL) == "" {
		return LoginWindow{}, errors.New("browser connection URL is required")
	}
	cdp, err := dialCDP(ctx, connectURL)
	if err != nil {
		return LoginWindow{}, err
	}
	defer func() { _ = cdp.Close() }()
	pageSession, err := cdp.createPage(deepSeekSignInURL)
	if err != nil {
		return LoginWindow{}, err
	}
	var live struct {
		LiveURL string `json:"liveURL"`
	}
	if err := cdp.call("Browserless.liveURL", map[string]any{
		"timeout":              int64((10 * time.Minute) / time.Millisecond),
		"interactable":         true,
		"resizable":            true,
		"showBrowserInterface": true,
		"quality":              40,
	}, pageSession, &live); err != nil {
		return LoginWindow{}, err
	}
	if strings.TrimSpace(live.LiveURL) == "" {
		return LoginWindow{}, errors.New("browser provider did not return an interactive URL")
	}
	return LoginWindow{URL: live.LiveURL}, nil
}

func (s *Service) PageSession(connectURL string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cdp, err := dialCDP(ctx, connectURL)
	if err != nil {
		return "", err
	}
	defer func() { _ = cdp.Close() }()
	return cdp.findPageSession()
}

func ValidProfileID(value string) bool {
	return safeID(value)
}

func NewProfileID() (string, error) {
	return randomProfileID()
}
