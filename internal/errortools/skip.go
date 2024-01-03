// Package errortools defines a set of additional error handling utilities,
// such as skipping certain errors where otherwise LibRate would panic
package errortools

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
)

// @function ParseIgnorableErrors
// @brief ParseIgnorableErrors parses a comma-separated list of ignorable errors
// @param fromFlag *string - the flag to parse from (comma-separated list of accepted error codes)
// @return

func ParseIgnorableErrors(fromFlag *string) (errors []string, err error) {
	if fromFlag == nil || *fromFlag == "" {
		return nil, nil
	}

	errorsList := strings.Split(*fromFlag, ",")
	errors = make([]string, len(errorsList))

	for i, err := range errorsList {
		errors[i] = strings.TrimSpace(err)
	}

	if !lo.Some(errorsList, Codes) {
		wrogCodes := lo.Reject(errorsList, func(code string, _ int) bool {
			return lo.Contains(Codes, code)
		})
		return nil, fmt.Errorf(
			"invalid ignorable error code %s specified. Acceptable error codes: %v",
			wrogCodes, Codes)
	}

	return errorsList, nil
}
