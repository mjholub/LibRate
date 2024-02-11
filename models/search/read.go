package searchdb

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-kivik/kivik/v4"
)

// we only care about scanning the data into it's raw JSON representation
// so that we can then use it for indexing
func (s *Storage) ReadAll(ctx context.Context, target TargetDB) (data []interface{}, err error) {
	if ok, err := s.client.DBExists(ctx, target.String()); !ok {
		return nil, fmt.Errorf("failed to check if database exists: %w", err)
	}
	db := s.client.DB(target.String())

	rows := db.AllDocs(ctx, nil)
	defer rows.Close()

	// can't merge []byte directly, so we'll still need to marshal that
	// to JSON again :(
	err = kivik.ScanAllDocs(rows, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to scan all documents: %w", err)
	}

	return data, nil
}

// ReadNew handles the cyclical or continuous
// retrieval of new documents for indexing.
// This function does not accept a target parameter, since it doesn't
// make much sense to only look for partial data.
// The channel parameter here is used to communicate data
// to the search indexer
func (s *Storage) ReadNew(
	ctx context.Context,
	outgoingDataFeed chan interface{},
	continuousMode bool,
) (err error) {
	// spawn workers for every db
	var wg sync.WaitGroup
	errorCh := make(chan error)
	for _, target := range AllTargets {
		db := s.client.DB(target.String())
		wg.Add(1)
		go s.readNewWorker(ctx, db, &wg, errorCh, outgoingDataFeed, continuousMode)
	}
	wg.Wait()

	select {
	case err = <-errorCh:
		return err
	default:
		return nil
	}
}

// readNewWorker is a worker that continuously
// updates the target changes feed to a data channel
// until explicitly stopped.
func (s *Storage) readNewWorker(
	ctx context.Context,
	target *kivik.DB,
	wg *sync.WaitGroup,
	errorCh chan error,
	outgoingDataFeed chan interface{},
	continuousMode bool,
) {
	select {
	case <-ctx.Done():
		return
	default:
		defer wg.Done()

		changes := target.Changes(ctx)
		data := make([]interface{}, 0)
		if !continuousMode {
			for changes.Next() {
				if err := changes.ScanDoc(data); err != nil {
					errorCh <- fmt.Errorf("failed to scan changes for %s: %w", target.Name(), err)
					return
				}
				outgoingDataFeed <- data
			}
			if err := changes.Err(); err != nil {
				errorCh <- fmt.Errorf("failed to get changes for %s: %w", target.Name(), err)
			}
			return
		}
		if err := changes.ScanDoc(data); err != nil {
			s.log.Error().Err(err).Msgf("failed to scan changes in goroutine for %s", target.Name())
		}
		outgoingDataFeed <- data
	}
}
