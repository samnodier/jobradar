package convert

import (
	"time"
)

func ToNullInt32(val int) *int32 {
	if val == 0 {
		return nil
	}
	v := int32(val)
	return &v
}

func ToNullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ToNullTime(val time.Time) *time.Time {
	if val.IsZero() {
		return nil
	}
	return &val
}
