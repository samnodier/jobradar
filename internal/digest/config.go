package digest

import (
	"fmt"
	"github.com/samnodier/jobradar/internal/mailer"
	"net/mail"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Enabled                   bool
	Recipient, From, Provider string
	Location                  *time.Location
	Hour, MinScore, MaxJobs   int
}

func FromEnv() (Config, mailer.Sender, error) {
	c := Config{Enabled: os.Getenv("DIGEST_ENABLED") == "true", Recipient: os.Getenv("DIGEST_RECIPIENT"), From: os.Getenv("EMAIL_FROM"), Provider: os.Getenv("EMAIL_PROVIDER"), Hour: 8, MinScore: 50, MaxJobs: 10}
	zone := os.Getenv("DIGEST_TIMEZONE")
	if zone == "" {
		zone = "America/New_York"
	}
	var err error
	c.Location, err = time.LoadLocation(zone)
	if err != nil {
		return c, nil, fmt.Errorf("invalid DIGEST_TIMEZONE")
	}
	for _, item := range []struct {
		name     string
		ptr      *int
		min, max int
	}{{"DIGEST_HOUR", &c.Hour, 0, 23}, {"DIGEST_MIN_SCORE", &c.MinScore, 0, 100}, {"DIGEST_MAX_JOBS", &c.MaxJobs, 1, 25}} {
		if val := os.Getenv(item.name); val != "" {
			n, e := strconv.Atoi(val)
			if e != nil || n < item.min || n > item.max {
				return c, nil, fmt.Errorf("invalid %s", item.name)
			}
			*item.ptr = n
		}
	}
	if !c.Enabled {
		return c, nil, nil
	}
	for _, address := range []string{c.Recipient, c.From} {
		if _, err := mail.ParseAddress(address); err != nil {
			return c, nil, fmt.Errorf("DIGEST_RECIPIENT and EMAIL_FROM must be valid addresses")
		}
	}
	recipient, _ := mail.ParseAddress(c.Recipient)
	c.Recipient = recipient.Address
	var sender mailer.Sender
	switch c.Provider {
	case "resend":
		if os.Getenv("RESEND_API_KEY") == "" {
			return c, nil, fmt.Errorf("RESEND_API_KEY required")
		}
		sender = &mailer.Resend{APIKey: os.Getenv("RESEND_API_KEY")}
	case "gohighlevel":
		if os.Getenv("GHL_API_KEY") == "" || os.Getenv("GHL_CONTACT_ID") == "" {
			return c, nil, fmt.Errorf("GHL_API_KEY and GHL_CONTACT_ID required")
		}
		sender = &mailer.GoHighLevel{APIKey: os.Getenv("GHL_API_KEY"), ContactID: os.Getenv("GHL_CONTACT_ID")}
	default:
		return c, nil, fmt.Errorf("EMAIL_PROVIDER must be resend or gohighlevel")
	}
	return c, sender, nil
}
