package searchdb

import (
	"context"
	"fmt"

	"github.com/goccy/go-json"
)

// we only care about scanning the data into it's raw JSON representation
// so that we can then use it for indexing
func (s *Storage) ReadAll(ctx context.Context, target TargetDB) (data []byte, err error) {
	if ok, err := s.client.DBExists(ctx, target.String()); !ok {
		return nil, fmt.Errorf("failed to check if database exists: %w", err)
	}
	db := s.client.DB(target.String())

	rows := db.AllDocs(ctx, nil)
	defer rows.Close()

	// can't merge []byte directly, so we'll still need to marshal that
	// to JSON again :(
	var dataChunks []interface{}
	for rows.Next() {
		var doc interface{}
		if err := rows.ScanDoc(&doc); err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		dataChunks = append(dataChunks, doc)
	}

	return json.Marshal(dataChunks)
}
