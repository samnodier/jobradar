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
	GitHub        GitHubOAuthConfig
	AppBaseURL    string
	OnboardingTTL time.Duration
	IsProduction  bool
}

func (cfg *AuthConfig) Validate() error {
	if cfg.GitHub.ClientID == "" {
		return fmt.Errorf("missing github client id")
	}
	if cfg.GitHub.ClientSecret == "" {
		return fmt.Errorf("missing github client secret")
	}
	if cfg.GitHub.RedirectURL == "" {
		return fmt.Errorf("missing github redirect url")
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
