package errortools

const (
	// ErrSQLCipherParse is the error code for SQLCipher parse errors.
	ErrSQLCipherParse = "ERR_SQLCIPHER_PARSE"
)

// nolint:gochecknoglobals
var Codes = []string{ErrSQLCipherParse}
