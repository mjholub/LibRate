package searchdb

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	_ "github.com/go-kivik/couchdb/v3"
	"github.com/go-kivik/kivik/v3"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
)

type (
	// Those types are simplified representations of what is
	// stored in postgres that are written to couchdb on insert/update
	Genre struct {
		ID           string               `json:"_id" mapstructure:"-"`
		Rev          string               `json:"_rev" mapstructure:"-"`
		Name         string               `json:"name" mapstructure:"name"`
		Kinds        []string             `json:"kinds" mapstructure:"kinds"`
		Descriptions [][]GenreDescription `json:"descriptions" mapstructure:"descriptions"`
	}

	GenreDescription struct {
		Description string `json:"description" mapstructure:"description"`
		Language    string `json:"language" mapstructure:"language"`
	}

	Member struct {
		ID          string `json:"_id" mapstructure:"-"`
		Rev         string `json:"_rev" mapstructure:"-"`
		Bio         string `json:"bio,omitempty" mapstructure:"bio,omitempty"`
		Webfinger   string `json:"webfinger,omitempty" mapstructure:"webfinger,omitempty"`
		DisplayName string `json:"display_name,omitempty" mapstructure:"display_name,omitempty"`
	}

	Studio struct {
		ID       string    `json:"_id" mapstructure:"-"`
		Rev      string    `json:"_rev" mapstructure:"-"`
		Name     string    `json:"name" mapstructure:"name"`
		Kind     string    `json:"kind" mapstructure:"kind"`
		CityUUID string    `json:"city" mapstructure:"city"`
		Added    time.Time `json:"added" mapstructure:"added"`
		Modified time.Time `json:"modified" mapstructure:"modified"`
	}

	Rating struct {
		ID         string `json:"_id" mapstructure:"-"`
		Rev        string `json:"_rev" mapstructure:"-"`
		Topic      string `json:"topic" mapstructure:"topic"`
		Body       string `json:"body" mapstructure:"body"`
		User       string `json:"user" mapstructure:"user"`
		MediaTitle string `json:"media_title" mapstructure:"media_title"`
		// not sure whether this shouldn't actually be a string as well
		Added    time.Time `json:"added" mapstructure:"added"`
		Modified time.Time `json:"modified" mapstructure:"modified"`
	}

	Artist struct {
		ID        string    `json:"_id" mapstructure:"-"`
		Rev       string    `json:"_rev" mapstructure:"-"`
		Name      string    `json:"name" mapstructure:"name"`
		Nicknames []string  `json:"nick_names" mapstructure:"nick_names"`
		Bio       string    `json:"bio" mapstructure:"bio"`
		Added     time.Time `json:"added" mapstructure:"added"`
		Modified  time.Time `json:"modified" mapstructure:"modified"`
	}

	Media struct {
		ID    string `json:"_id" mapstructure:"-"`
		Rev   string `json:"_rev" mapstructure:"-"`
		Title string `json:"title" mapstructure:"title"`
		Kind  string `json:"kind" mapstructure:"kind"`
		// Created refers to the release date
		Created  time.Time `json:"created" mapstructure:"created"`
		Added    time.Time `json:"added" mapstructure:"added"`
		Modified time.Time `json:"modified" mapstructure:"modified"`
	}

	CombinedData struct {
		Genres  []Genre  `json:"genres" mapstructure:"genres"`
		Members []Member `json:"members" mapstructure:"members"`
		Studios []Studio `json:"studio" mapstructure:"studio"`
		Ratings []Rating `json:"ratings" mapstructure:"ratings"`
		Artists []Artist `json:"artists" mapstructure:"artists"`
		Media   []Media  `json:"media" mapstructure:"media"`
	}

	BleveDocument struct {
		// ID is the couchdb id of the document
		ID string `json:"id"`
		// Type is a field that allows us to distinguish between different types of documents
		Type string `json:"type"`
		// Fields is the list of additional fields that we can aggregate on
		Fields []interface{} `json:"fields"`
		// Data is the raw representation of the document
		Data map[string]interface{} `json:"data"`
	}

	// AnonymousDocument is same as Ble
	AnonymousDocument struct {
		Type   string                 `json:"type"`
		Fields []interface{}          `json:"fields"`
		Data   map[string]interface{} `json:"data"`
	}
)

// we only care about scanning the data into it's raw JSON representation
// so that we can then use it for indexing
func (s *Storage) ReadAll(ctx context.Context) (data *CombinedData, err error) {
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
		s.log.Debug().Msg("Finished reading from couchDB")
		return &combinedData, nil
	}
}

func ToBleveDocument(combinedData *CombinedData, log *zerolog.Logger) (docs []BleveDocument, err error) {
	errorCh := make(chan error, 1)
	var wg sync.WaitGroup
	genreDocCh := make(chan []BleveDocument, 1)
	wg.Add(1)
	go func(log *zerolog.Logger) {
		var genres []BleveDocument
		defer wg.Done()
		for i := range combinedData.Genres {
			var doc BleveDocument
			doc.ID = combinedData.Genres[i].ID
			log.Debug().Msgf("processing document with ID: %s (DB: genres)", doc.ID)
			doc.Type = "genres"
			doc.Fields = []interface{}{combinedData.Genres[i].Name, combinedData.Genres[i].Kinds, combinedData.Genres[i].Descriptions}
			err := mapstructure.Decode(combinedData.Genres[i], &doc.Data)
			if err != nil {
				log.Error().Err(err).Msgf("error converting struct into map (genres): %v", err)
				errorCh <- fmt.Errorf("error converting struct into map: %w", err)
				break
			}
			log.Trace().Msgf("document data: %+v", doc)
			genres = append(genres, doc)
		}
		genreDocCh <- genres
	}(log)

	wg.Add(1)
	memberDocCh := make(chan []BleveDocument, 1)
	go func() {
		defer wg.Done()
		var members []BleveDocument
		for i := range combinedData.Members {
			var doc BleveDocument
			doc.ID = combinedData.Members[i].ID
			log.Debug().Msgf("processing document with ID: %s (DB: members)", doc.ID)
			doc.Type = "members"
			doc.Fields = []interface{}{combinedData.Members[i].Bio, combinedData.Members[i].Webfinger, combinedData.Members[i].DisplayName}
			if err := mapstructure.Decode(combinedData.Members[i], &doc.Data); err != nil {
				errorCh <- fmt.Errorf("error converting struct into map: %w", err)
				break
			}
			members = append(members, doc)
		}
		memberDocCh <- members
	}()

	wg.Add(1)
	studioDocCh := make(chan []BleveDocument, 1)
	go func() {
		defer wg.Done()
		var studios []BleveDocument
		for i := range combinedData.Studios {
			var doc BleveDocument
			doc.ID = combinedData.Studios[i].ID
			doc.Type = "studios"

			doc.Fields = []interface{}{combinedData.Studios[i].Name, combinedData.Studios[i].Kind, combinedData.Studios[i].Added}
			if err := mapstructure.Decode(combinedData.Studios[i], &doc.Data); err != nil {
				errorCh <- fmt.Errorf("error converting struct into map: %w", err)
				break
			}
			studios = append(studios, doc)
		}
		studioDocCh <- studios
	}()

	wg.Add(1)
	reviewDocCh := make(chan []BleveDocument, 1)
	go func() {
		defer wg.Done()
		var ratings []BleveDocument
		for i := range combinedData.Ratings {
			var doc BleveDocument
			doc.ID = combinedData.Ratings[i].ID
			doc.Type = "ratings"

			doc.Fields = []interface{}{combinedData.Ratings[i].Topic, combinedData.Ratings[i].Body, combinedData.Ratings[i].User, combinedData.Ratings[i].MediaTitle, combinedData.Ratings[i].Added}
			if err := mapstructure.Decode(combinedData.Ratings[i], &doc.Data); err != nil {
				errorCh <- fmt.Errorf("error converting struct into map: %w", err)
				break
			}

			ratings = append(ratings, doc)
		}
		reviewDocCh <- ratings
	}()

	wg.Add(1)
	artistsDocCh := make(chan []BleveDocument, 1)
	go func() {
		defer wg.Done()
		var artists []BleveDocument
		for i := range combinedData.Artists {
			var doc BleveDocument
			doc.ID = combinedData.Artists[i].ID
			doc.Type = "artists"

			doc.Fields = []interface{}{combinedData.Artists[i].Name, combinedData.Artists[i].Nicknames, combinedData.Artists[i].Bio, combinedData.Artists[i].Added}
			if err := mapstructure.Decode(combinedData.Artists[i], &doc.Data); err != nil {
				errorCh <- fmt.Errorf("error converting struct into map: %w", err)
				break
			}
			artists = append(artists, doc)
		}
		artistsDocCh <- artists
	}()

	wg.Add(1)
	mediaDocCh := make(chan []BleveDocument, 1)
	go func() {
		defer wg.Done()
		var mediaDocs []BleveDocument
		for i := range combinedData.Media {
			var doc BleveDocument
			doc.ID = combinedData.Media[i].ID
			doc.Type = "media"

			doc.Fields = []interface{}{combinedData.Media[i].Title, combinedData.Media[i].Kind, combinedData.Media[i].Created, combinedData.Media[i].Added}
			if err := mapstructure.Decode(combinedData.Media[i], &doc.Data); err != nil {
				errorCh <- fmt.Errorf("error converting struct into map: %w", err)
				break
			}
			mediaDocs = append(mediaDocs, doc)
		}
		mediaDocCh <- mediaDocs
	}()

	wg.Wait()
	close(errorCh)

	// NOTE: we declare separate channels since we're using slices,
	// Therefore we need to collect slices for each DB first,
	// then merge them using slices.Concat
	close(mediaDocCh)
	close(artistsDocCh)
	close(memberDocCh)
	close(genreDocCh)
	close(studioDocCh)
	close(reviewDocCh)

	if err = <-errorCh; err != nil {
		return nil, err
	}
	mediaDocs := <-mediaDocCh
	artistDocs := <-artistsDocCh
	memberDocs := <-memberDocCh
	genreDocs := <-genreDocCh
	log.Debug().Msgf("genreDocs: %+v", genreDocs)
	studioDocs := <-studioDocCh
	reviewDocs := <-reviewDocCh
	docs = slices.Concat(mediaDocs, artistDocs, memberDocs, genreDocs, studioDocs, reviewDocs)
	return docs, nil
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
