package db

import (
	"database/sql"
	"time"
)

// IsNotNull returns true if a value is not "nullish"
// falsity is not considered nullish
func IsNotNull(value interface{}) bool {
	return value != nil && value != "" && value != 0 && value != sql.NullInt64{
		Int64: 0,
		Valid: false,
	} && value != sql.NullString{
		String: "",
		Valid:  false,
	} && value != sql.NullBool{
		Bool:  false,
		Valid: false,
	} && value != sql.NullFloat64{
		Float64: 0,
		Valid:   false,
	} && value != sql.NullTime{
		Time:  time.Time{},
		Valid: false,
	} && value != sql.NullInt32{
		Int32: 0,
		Valid: false,
	} && value != sql.NullInt16{
		Int16: 0,
		Valid: false,
	} && value != sql.NullByte{
		Byte:  0,
		Valid: false,
	}
}
