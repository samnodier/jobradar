package auth

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/testutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestReadTokenScopeOwnershipAndRevocation(t *testing.T) {
	p := testutil.Pool(t)
	ctx := context.Background()
	q := database.New(p)
	id := uuid.New()
	email := id.String() + "@example.test"
	if _, err := p.Exec(ctx, "INSERT INTO users(id,email,username)VALUES($1,$2::text,$2::text)", id, email); err != nil {
		t.Fatal(err)
	}
	defer p.Exec(ctx, "DELETE FROM users WHERE id=$1", id)
	token, hash, err := NewReadToken()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(hash, token) || len(token) != 67 {
		t.Fatal("invalid token storage")
	}
	row, err := q.CreateReadToken(ctx, database.CreateReadTokenParams{UserID: id, Name: "test", TokenHash: hash, ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true}})
	if err != nil {
		t.Fatal(err)
	}
	h := &Handler{db: q}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, ok := SessionFromContext(r.Context())
		if !ok || session.UserID != id {
			t.Error("wrong account")
		}
		w.WriteHeader(204)
	})
	call := func(method, secret string) int {
		req := httptest.NewRequest(method, "/export/jobs", nil)
		req.Header.Set("Authorization", "Bearer "+secret)
		w := httptest.NewRecorder()
		h.RequireReadAccess(next).ServeHTTP(w, req)
		return w.Code
	}
	if call("GET", token) != 204 || call("POST", token) != 405 || call("GET", token+"bad") != 401 {
		t.Fatal("scope/auth failure")
	}
	h.cfg.PersonalEmail = "another@example.test"
	if call("GET", token) != 403 {
		t.Fatal("personal restriction bypass")
	}
	h.cfg.PersonalEmail = ""
	if _, err = p.Exec(ctx, "UPDATE read_api_tokens SET expires_at=now()-interval '1 minute' WHERE id=$1", row.ID); err != nil {
		t.Fatal(err)
	}
	if call("GET", token) != 401 {
		t.Fatal("expired token accepted")
	}
	p.Exec(ctx, "UPDATE read_api_tokens SET expires_at=now()+interval '1 hour' WHERE id=$1", row.ID)
	n, err := q.RevokeReadToken(ctx, database.RevokeReadTokenParams{ID: row.ID, UserID: uuid.New()})
	if err != nil || n != 0 {
		t.Fatal("cross-account revocation")
	}
	if _, err = q.RevokeReadToken(ctx, database.RevokeReadTokenParams{ID: row.ID, UserID: id}); err != nil {
		t.Fatal(err)
	}
	if call("GET", token) != 401 {
		t.Fatal("revoked token accepted")
	}
}
