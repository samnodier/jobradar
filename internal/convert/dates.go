package convert

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// ToDateRequired parses a YYYY-MM-DD string into a pgtype.Date.
// Returns an error if the string is empty or malformed.
func ToDateRequired(s string) (pgtype.Date, error) {
	if s == "" {
		return pgtype.Date{}, fmt.Errorf("date is required")
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return pgtype.Date{}, fmt.Errorf("invalid start_date format, expected YYYY-MM-DD")
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}

// ToDateOptional parses a YYYY-MM-DD string into a pgtype.Date.
// Returns an invalid (NULL) pgtype.Date if the string is empty.
func ToDateOptional(s string) (pgtype.Date, error) {
	if s == "" {
		return pgtype.Date{Valid: false}, nil // NULL in DB
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return pgtype.Date{}, fmt.Errorf("invalid date format, expected YYYY-MM-DD")
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}
