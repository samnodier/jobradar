package convert

import (
	"time"
)

func ToNullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func FromNullString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ToNullInt32(val int) *int32 {
	v := int32(val)
	return &v
}

func ToNullTime(val time.Time) *time.Time {
	if val.IsZero() {
		return nil
	}
	return &val
}
