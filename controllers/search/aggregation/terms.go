package aggregation

import "time"

type termsAggregator struct {
	// terms is a map of terms to their counts
	terms map[string]uint
}

type dateAggregator struct {
	Since time.Time
	Till  time.Time
}
