package searchdb

import (
	"context"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-kivik/couchdb/v3"
	"github.com/go-kivik/kivik/v3"
	"github.com/goccy/go-json"
)

type (
	// Those types are simplified representations of what is
	// stored in postgres that are written to couchdb on insert/update
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

	CombinedData struct {
		Genres  []Genre  `json:"genres"`
		Members []Member `json:"users"`
		Studios []Studio `json:"studios"`
		Ratings []Rating `json:"ratings"`
		Artists []Artist `json:"artists"`
		Media   []Media  `json:"media"`
	}
)

// we only care about scanning the data into it's raw JSON representation
// so that we can then use it for indexing
func (s *Storage) ReadAll(ctx context.Context) (data []byte, err error) {
	var wg sync.WaitGroup

	wg.Add(1)
	genresCh := make(chan []Genre, 1)
	errorCh := make(chan error, 1)
	go func(ctx context.Context) {
		defer wg.Done()
		data, err := s.ReadGenres(ctx)
		if err != nil {
			errorCh <- err
			return
		}
		genresCh <- data
	}(ctx)

	wg.Add(1)
	membersCh := make(chan []Member, 1)
	go func(ctx context.Context) {
		defer wg.Done()
		data, err := s.ReadMembers(ctx)
		if err != nil {
			errorCh <- err
			return
		}
		membersCh <- data
	}(ctx)

	wg.Add(1)
	studioCh := make(chan []Studio, 1)
	go func(ctx context.Context) {
		defer wg.Done()
		data, err := s.ReadStudios(ctx)
		if err != nil {
			errorCh <- err
			return
		}
		studioCh <- data
	}(ctx)

	wg.Add(1)
	reviewsCh := make(chan []Rating, 1)
	go func(ctx context.Context) {
		defer wg.Done()
		data, err := s.ReadRatings(ctx)
		if err != nil {
			errorCh <- err
			return
		}
		reviewsCh <- data
	}(ctx)

	wg.Add(1)
	artistsCh := make(chan []Artist, 1)
	go func(ctx context.Context) {
		defer wg.Done()
		data, err := s.ReadArtists(ctx)
		if err != nil {
			errorCh <- err
			return
		}
		artistsCh <- data
	}(ctx)

	wg.Add(1)
	mediaCh := make(chan []Media, 1)
	go func(ctx context.Context) {
		defer wg.Done()
		data, err := s.ReadMedia(ctx)
		if err != nil {
			errorCh <- err
			return
		}
		mediaCh <- data
	}(ctx)

	wg.Wait()
	select {
	case err = <-errorCh:
		return nil, err
	default:
		close(genresCh)
		close(membersCh)
		close(studioCh)
		close(reviewsCh)
		close(artistsCh)
		close(mediaCh)
		combinedData := CombinedData{
			Genres:  <-genresCh,
			Members: <-membersCh,
			Studios: <-studioCh,
			Ratings: <-reviewsCh,
			Artists: <-artistsCh,
			Media:   <-mediaCh,
		}
		return json.Marshal(combinedData)
	}
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
