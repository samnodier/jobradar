package auth

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samnodier/jobradar/internal/database"
)

// IssuePasswordSetup is a server-owner operation, NEVER a public HTTP handler.
// Possession of the database credentials authorizes local provisioning/recovery.
func IssuePasswordSetup(ctx context.Context, pool *pgxpool.Pool, email string) (string, error) {
	email, err := NormalizeEmail(email)
	if err != nil {
		return "", err
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)
	q := database.New(tx)
	user, err := q.GetAuthIdentityByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		created, e := q.CreateUser(ctx, database.CreateUserParams{Email: email, Username: "local_" + uuid.NewString()[:12]})
		if e != nil {
			return "", e
		}
		user.ID = created.ID
	} else if err != nil {
		return "", err
	}
	if _, err = q.LockAuthIdentity(ctx, user.ID); err != nil {
		return "", err
	}
	token, hash, err := NewReadToken()
	if err != nil {
		return "", err
	}
	if err = q.DeletePasswordSetupTokens(ctx, user.ID); err != nil {
		return "", err
	}
	if err = q.CreatePasswordSetupToken(ctx, database.CreatePasswordSetupTokenParams{TokenHash: hash, UserID: user.ID}); err != nil {
		return "", err
	}
	if err = tx.Commit(ctx); err != nil {
		return "", err
	}
	return token, nil
}
