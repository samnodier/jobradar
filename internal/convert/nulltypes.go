package convert

import (
	"database/sql"
)

func ToNullInt32(val int) sql.NullInt32 {
	if val == 0 {
		return sql.NullInt32{
			Valid: false,
		}
	}
	return sql.NullInt32{
		Int32: int32(val),
		Valid: true,
	}
}

func ToNullString(val string) sql.NullString {
	if val == "" {
		return sql.NullString{
			Valid: false,
		}
	}
	return sql.NullString{
		String: val,
		Valid:  true,
	}
}

func ToNullBool(val *bool) sql.NullBool {
	if val == nil {
		return sql.NullBool{
			Valid: false,
		}
	}
	return sql.NullBool{Bool: *val, Valid: true}
}
