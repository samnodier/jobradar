package auth

import (
	"fmt"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type GitHubOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}
type AuthConfig struct {
	GitHub         GitHubOAuthConfig
	AppBaseURL     string
	OnboardingTTL  time.Duration
	SessionTTL     time.Duration
	IsProduction   bool
	PersonalEmail  string
	AllowedOrigins []string
}

func (cfg *AuthConfig) Validate() error {
	if cfg.GitHub.ClientID != "" || cfg.GitHub.ClientSecret != "" {
		if cfg.GitHub.ClientID == "" || cfg.GitHub.ClientSecret == "" || cfg.GitHub.RedirectURL == "" {
			return fmt.Errorf("GitHub requires client ID, secret and callback URL when enabled")
		}
	}
	if cfg.SessionTTL == 0 {
		cfg.SessionTTL = 7 * 24 * time.Hour
	}
	if cfg.SessionTTL < 0 {
		return fmt.Errorf("invalid session ttl")
	}
	if cfg.AppBaseURL == "" {
		return fmt.Errorf("missing app base url")
	}
	if cfg.OnboardingTTL == 0 {
		cfg.OnboardingTTL = 10 * time.Minute
	}
	if cfg.OnboardingTTL < 0 {
		return fmt.Errorf("invalid onboarding ttl")
	}
	return nil
}

func (cfg AuthConfig) GitHubOAuth() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.GitHub.ClientID,
		ClientSecret: cfg.GitHub.ClientSecret,
		RedirectURL:  cfg.GitHub.RedirectURL,
		Scopes: []string{
			"read:user",
			"user:email",
		},
		Endpoint: github.Endpoint,
	}
}

func (cfg AuthConfig) GitHubEnabled() bool {
	return cfg.GitHub.ClientID != "" && cfg.GitHub.ClientSecret != ""
}
