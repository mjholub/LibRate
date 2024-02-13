package searchdb

import (
	"context"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-kivik/couchdb/v3"
	"github.com/go-kivik/kivik/v3"
)

type (
	// Those types are simplified representations of what is
	// stored in postgres that are written on insert/update
	// to couchdb
	Genre struct {
		ID           string               `json:"_id"`
		Rev          string               `json:"_rev"`
		Name         string               `json:"name"`
		Kinds        []string             `json:"kinds"`
		Descriptions [][]GenreDescription `json:"descriptions"`
	}

	GenreDescription struct {
		Description string `json:"description"`
		Language    string `json:"language"`
	}

	Member struct {
		ID          string `json:"_id"`
		Rev         string `json:"_rev"`
		Bio         string `json:"bio,omitempty"`
		Webfinger   string `json:"webfinger,omitempty"`
		DisplayName string `json:"display_name,omitempty"`
	}

	Studio struct {
		ID       string    `json:"_id"`
		Rev      string    `json:"_rev"`
		Name     string    `json:"name"`
		Kind     string    `json:"kind"`
		CityUUID string    `json:"city"`
		Added    time.Time `json:"added"`
		Modified time.Time `json:"modified"`
	}

	Rating struct {
		ID         string `json:"_id"`
		Rev        string `json:"_rev"`
		Topic      string `json:"topic"`
		Body       string `json:"body"`
		User       string `json:"user"`
		MediaTitle string `json:"media_title"`
		// not sure whether this shouldn't actually be a string as well
		Added    time.Time `json:"added"`
		Modified time.Time `json:"modified"`
	}

	Artist struct {
		ID        string    `json:"_id"`
		Rev       string    `json:"_rev"`
		Name      string    `json:"name"`
		Nicknames string    `json:"nick_names"`
		Bio       string    `json:"bio"`
		Added     time.Time `json:"added"`
		Modified  time.Time `json:"modified"`
	}

	Media struct {
		ID    string `json:"_id"`
		Rev   string `json:"_rev"`
		Title string `json:"title"`
		Kind  string `json:"kind"`
		// Created refers to the release date
		Created  time.Time `json:"created"`
		Added    time.Time `json:"added"`
		Modified time.Time `json:"modified"`
	}
)

// we only care about scanning the data into it's raw JSON representation
// so that we can then use it for indexing
func (s *Storage) ReadAll(ctx context.Context, target TargetDB) (data []interface{}, err error) {
	if ok, err := s.client.DBExists(ctx, target.String()); !ok {
		return nil, fmt.Errorf("failed to check if database exists: %w", err)
	}
	db := s.client.DB(ctx, target.String())

	options := map[string]interface{}{
		"include_docs": true,
	}

	rows, err := db.AllDocs(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("error accessing rows: %v", err)
	}
	defer rows.Close()

	// can't merge []byte directly, so we'll still need to marshal that
	// to JSON again :(
	for rows.Next() {
		err = rows.ScanDoc(&data)
		if err != nil {
			return nil, fmt.Errorf("failed to scan all documents: %w", err)
		}
	}

	return data, nil
}

func (s *Storage) ReadGenres(ctx context.Context) (data []Genre, err error) {
	if ok, err := s.client.DBExists(ctx, Genres.String()); !ok {
		return nil, fmt.Errorf("failed to check if database exists: %w", err)
	}
	db := s.client.DB(ctx, Genres.String())

	options := map[string]interface{}{
		"include_docs": true,
	}

	rows, err := db.AllDocs(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("error accessing database rows: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var genre Genre
		err = rows.ScanDoc(&genre)
		if err != nil {
			return nil, fmt.Errorf("failed to scan all documents: %w", err)
		}
		data = append(data, genre)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("errors found in the result set: %w", rows.Err())
	}

	return data, nil
}

func (s *Storage) ReadMembers(ctx context.Context) (data []Member, err error) {
	if ok, err := s.client.DBExists(ctx, Members.String()); !ok {
		return nil, fmt.Errorf("failed to check if database exists: %w", err)
	}
	db := s.client.DB(ctx, Members.String())

	options := map[string]interface{}{
		"include_docs": true,
	}
	rows, err := db.AllDocs(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("error accessing database rows: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var member Member
		err = rows.ScanDoc(&member)
		if err != nil {
			return nil, fmt.Errorf("failed to scan all documents: %w", err)
		}
		data = append(data, member)
	}

	return data, nil
}

func (s *Storage) ReadStudios(ctx context.Context) (data []Studio, err error) {
	if ok, err := s.client.DBExists(ctx, Studios.String()); !ok {
		return nil, fmt.Errorf("failed to check if database exists: %w", err)
	}

	db := s.client.DB(ctx, Studios.String())

	options := map[string]interface{}{
		"include_docs": true,
	}
	rows, err := db.AllDocs(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("error accessing database rows: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var studio Studio
		err = rows.ScanDoc(&studio)
		if err != nil {
			return nil, fmt.Errorf("failed to scan all documents: %w", err)
		}
		data = append(data, studio)
	}

	return data, nil
}

func (s *Storage) ReadRatings(ctx context.Context) (data []Rating, err error) {
	if ok, err := s.client.DBExists(ctx, Ratings.String()); !ok {
		return nil, fmt.Errorf("failed to check if database exists: %w", err)
	}

	db := s.client.DB(ctx, Ratings.String())

	options := map[string]interface{}{
		"include_docs": true,
	}
	rows, err := db.AllDocs(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("error accessing database rows: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rating Rating
		err = rows.ScanDoc(&rating)
		if err != nil {
			return nil, fmt.Errorf("failed to scan all documents: %w", err)
		}
		data = append(data, rating)
	}

	return data, nil
}

func (s *Storage) ReadArtists(ctx context.Context) (data []Artist, err error) {
	if ok, err := s.client.DBExists(ctx, Artists.String()); !ok {
		return nil, fmt.Errorf("failed to check if database exists: %w", err)
	}

	db := s.client.DB(ctx, Artists.String())

	options := map[string]interface{}{
		"include_docs": true,
	}
	rows, err := db.AllDocs(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("error accessing database rows: %v", err)
	}

	defer rows.Close()
	for rows.Next() {
		var artist Artist
		if err := rows.ScanDoc(&artist); err != nil {
			return nil, fmt.Errorf("failed to scan all documents: %w", err)
		}
		data = append(data, artist)
	}

	return data, nil
}

func (s *Storage) ReadMedia(ctx context.Context) (data []Media, err error) {
	if ok, err := s.client.DBExists(ctx, MediaDB.String()); !ok {
		return nil, fmt.Errorf("failed to check if database exists: %w", err)
	}

	db := s.client.DB(ctx, MediaDB.String())

	options := map[string]interface{}{
		"include_docs": true,
	}
	rows, err := db.AllDocs(ctx, options)
	defer rows.Close()

	for rows.Next() {
		var media Media
		err = rows.ScanDoc(&media)
		if err != nil {
			return nil, fmt.Errorf("failed to scan all documents: %w", err)
		}
		data = append(data, media)
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
		db := s.client.DB(ctx, target.String())
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

		changes, err := target.Changes(ctx)
		if err != nil {
			errorCh <- fmt.Errorf("failed to initialize database changes watcher: %v", err)
			return
		}
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
