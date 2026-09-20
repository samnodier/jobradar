// Package mailer sends a single transactional email; scheduling and persistence live elsewhere.
package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Message struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}
type Sender interface {
	Send(context.Context, string, Message) (string, error)
	RetrySafe() bool
}
type SendError struct {
	Code      int
	Ambiguous bool
}

func (e *SendError) Error() string {
	if e.Code == 0 {
		return "email provider connection failed; delivery status is unknown"
	}
	return fmt.Sprintf("email provider returned HTTP %d", e.Code)
}

type Resend struct {
	APIKey, BaseURL string
	Client          *http.Client
}

func (r *Resend) RetrySafe() bool { return true }
func (r *Resend) Send(ctx context.Context, key string, m Message) (string, error) {
	endpoint := r.BaseURL
	if endpoint == "" {
		endpoint = "https://api.resend.com"
	}
	return send(ctx, r.Client, endpoint+"/emails", r.APIKey, key, "", m, "id")
}

type GoHighLevel struct {
	APIKey, ContactID, BaseURL string
	Client                     *http.Client
}

func (g *GoHighLevel) RetrySafe() bool { return false }
func (g *GoHighLevel) Send(ctx context.Context, key string, m Message) (string, error) {
	endpoint := g.BaseURL
	if endpoint == "" {
		endpoint = "https://services.leadconnectorhq.com"
	}
	if len(m.To) != 1 {
		return "", &SendError{Code: 400}
	}
	client := g.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"/contacts/"+url.PathEscape(g.ContactID), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+g.APIKey)
	req.Header.Set("Version", "v3")
	resp, err := client.Do(req)
	if err != nil {
		return "", &SendError{Code: 429}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", &SendError{Code: resp.StatusCode}
	}
	var contact struct {
		Contact struct {
			Email string `json:"email"`
		} `json:"contact"`
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&contact); err != nil {
		return "", &SendError{Code: 400}
	}
	if !strings.EqualFold(contact.Contact.Email, m.To[0]) {
		return "", fmt.Errorf("GoHighLevel contact does not match DIGEST_RECIPIENT")
	}
	body := map[string]string{"type": "Email", "contactId": g.ContactID, "emailFrom": m.From, "emailTo": m.To[0], "subject": m.Subject, "html": m.HTML, "status": "pending"}
	return send(ctx, client, endpoint+"/conversations/messages", g.APIKey, "", "v3", body, "messageId")
}

func send(ctx context.Context, client *http.Client, endpoint, token, key, version string, body any, idField string) (string, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	if key != "" {
		req.Header.Set("Idempotency-Key", key)
	}
	if version != "" {
		req.Header.Set("Version", version)
	}
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", &SendError{Ambiguous: true}
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", &SendError{Code: resp.StatusCode, Ambiguous: resp.StatusCode >= 500}
	}
	var result map[string]json.RawMessage
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result); err != nil {
		return "", &SendError{Ambiguous: true}
	}
	var id string
	if err := json.Unmarshal(result[idField], &id); err != nil || id == "" {
		return "", &SendError{Ambiguous: true}
	}
	return id, nil
}
