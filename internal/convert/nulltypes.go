// Package convert...
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

func FromNullInt32(val *int32) int32 {
	if val == nil {
		return 0
	}
	return *val
}

func ToNullInt32(val int) *int32 {
	v := int32(val)
	return &v
}

func FromNullTime(val *time.Time) string {
	if val == nil {
		return ""
	}
	return val.Format(time.RFC3339)
}

func ToNullTime(val time.Time) *time.Time {
	if val.IsZero() {
		return nil
	}
	return &val
}
