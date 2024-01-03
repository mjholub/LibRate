package errortools

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseIgnorableErrors(t *testing.T) {
	tests := []struct {
		name          string
		fromFlag      *string
		want          []string
		expectedError error
	}{
		{
			name:          "empty flag",
			fromFlag:      stringPtr(""),
			want:          nil,
			expectedError: nil,
		},
		{
			name:          "nil flag",
			fromFlag:      nil,
			want:          nil,
			expectedError: nil,
		},
		{
			name:          "valid flag",
			fromFlag:      stringPtr("ERR_SQLCIPHER_PARSE"),
			want:          []string{"ERR_SQLCIPHER_PARSE"},
			expectedError: nil,
		},
		{
			name:          "invalid code literal",
			fromFlag:      stringPtr("thggf"),
			want:          nil,
			expectedError: fmt.Errorf("invalid ignorable error code [thggf] specified. Acceptable error codes: %v", Codes),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParseIgnorableErrors(test.fromFlag)
			assert.Equal(t, test.want, got)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
