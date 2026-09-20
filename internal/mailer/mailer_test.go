package mailer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResendIdempotencyAndFailureClassification(t *testing.T) {
	count := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		if r.Header.Get("Idempotency-Key") != "stable-key" {
			t.Error("missing stable idempotency key")
		}
		if count == 1 {
			w.Write([]byte(`{"id":"email-1"}`))
		} else {
			w.WriteHeader(503)
		}
	}))
	defer server.Close()
	sender := &Resend{APIKey: "secret", BaseURL: server.URL}
	m := Message{To: []string{"sam@example.test"}, Subject: "Jobs"}
	id, err := sender.Send(context.Background(), "stable-key", m)
	if err != nil || id != "email-1" {
		t.Fatal(id, err)
	}
	_, err = sender.Send(context.Background(), "stable-key", m)
	e, ok := err.(*SendError)
	if !ok || !e.Ambiguous {
		t.Fatalf("503 must be treated as potentially delivered: %v", err)
	}
}
func TestGHLRejectsWrongRecipientBeforeSend(t *testing.T) {
	posts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			posts++
		}
		w.Write([]byte(`{"contact":{"email":"other@example.test"}}`))
	}))
	defer server.Close()
	sender := &GoHighLevel{APIKey: "secret", ContactID: "contact", BaseURL: server.URL}
	_, err := sender.Send(context.Background(), "key", Message{To: []string{"sam@example.test"}})
	if err == nil || posts != 0 {
		t.Fatalf("wrong recipient must not receive mail: %v posts=%d", err, posts)
	}
}
