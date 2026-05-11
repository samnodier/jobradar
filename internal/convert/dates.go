package convert

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// ToDateRequired parses a YYYY-MM-DD or YYYY-MM string into a pgtype.Date.
// Returns an error if the string is empty or malformed.
func ToDateRequired(s string) (pgtype.Date, error) {
	if s == "" {
		return pgtype.Date{}, fmt.Errorf("date is required")
	}
	if len(s) == 7 { // YYYY-MM format from <input type="month">
		s = s + "-01"
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return pgtype.Date{Valid: false}, fmt.Errorf("invalid date format, expected YYYY-MM-DD or YYYY-MM")
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}

// ToDateOptional parses a YYYY-MM-DD or YYYY-MM string into a pgtype.Date.
// Returns an invalid (NULL) pgtype.Date if the string is empty.
func ToDateOptional(s string) (pgtype.Date, error) {
	if s == "" {
		return pgtype.Date{Valid: false}, nil // NULL in DB
	}
	if len(s) == 7 { // YYYY-MM format
		s = s + "-01"
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return pgtype.Date{Valid: false}, fmt.Errorf("invalid date format, expected YYYY-MM-DD or YYYY-MM")
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}
