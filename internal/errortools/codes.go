package errortools

const (
	// ErrSQLCipherParse is the error code for SQLCipher parse errors.
	ErrSQLCipherParse = "ERR_SQLCIPHER_PARSE"
	IOError           = "IOError"
)

// nolint:gochecknoglobals
var Codes = []string{ErrSQLCipherParse, IOError}
