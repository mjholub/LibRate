package media

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	scn "github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

type (
	Album struct {
		MediaID      *uuid.UUID     `json:"media_id" db:"media_id,pk,unique"`
		Name         string         `json:"name" db:"name"`
		AlbumArtists []AlbumArtist  `json:"album_artists" db:"album_artists"`
		ImagePaths   pq.StringArray `json:"image_paths,omitempty"` // we make use of a junction table that utilizes the image IDs
		ReleaseDate  time.Time      `json:"release_date" db:"release_date"`
		Genres       []Genre        `json:"genres,omitempty" db:"genres"`
		//	Studio       Studio                       `json:"studio,omitempty" db:"studio"`
		Keywords []Keyword    `json:"keywords,omitempty" db:"keywords"`
		Duration sql.NullTime `json:"duration,omitempty" db:"duration"`
		Tracks   []Track      `json:"tracks,omitempty" db:"tracks"`
		//	Languages int16         `json:"languages" db:"languages,omitempty"`
	}

	// junction table media.album_artists
	AlbumArtist struct {
		ID         uuid.UUID `json:"artist" db:"artist,pk,unique"`
		Name       string    `json:"name" db:"-"` // must perform a join operation to get the name
		ArtistType string    `json:"artist_type" db:"artist_type" validate:"required,oneof=individual group"`
	}

	Track struct {
		MediaID *uuid.UUID `json:"media_id" db:"media_id,pk,unique"`
		Name    string     `json:"name" db:"name"`
		AlbumID *uuid.UUID `json:"album_id" db:"album"`
		//		Artists   mo.Either[[]Person, []Group] `json:"artists" db:"artists"`
		Duration time.Time `json:"duration" db:"duration"`
		Lyrics   string    `json:"lyrics,omitempty" db:"lyrics"`
		Number   int16     `json:"track_number" db:"track_number"`
		// Languages []string                     `json:"languages,omitempty" db:"languages"`
	}
)

func addAlbum(ctx context.Context, db *pgxpool.Pool, album *Album) error {
	// Insert the album into the media.albums table
	_, err := db.Exec(ctx, `
		INSERT INTO media.albums (media_id, name, release_date, duration)
		VALUES ($1, $2, $3, $4)`,
		album.MediaID, album.Name, album.ReleaseDate, album.Duration)
	if err != nil {
		return fmt.Errorf("failed to insert album into media.albums: %w", err)
	}
	// Insert keywords into the junction table
	_, err = db.Exec(ctx, `
		INSERT INTO media.album_keywords (album, keyword)
		VALUES ($1, $2)`,
		album.MediaID, album.Keywords)
	if err != nil {
		return fmt.Errorf("failed to insert album keywords into media.album_keywords: %w", err)
	}

	// Insert the genres into the media.album_genres table
	for i := range album.Genres {
		_, err = db.Exec(ctx, "INSERT INTO media.album_genres (album, genre) VALUES ($1, $2)",
			album.MediaID, album.Genres[i].ID)
		if err != nil {
			return fmt.Errorf("failed to insert album genre into media.album_genres: %w", err)
		}
	}

	errChan := make(chan error)
	go func() {
		aa := album.AlbumArtists
		for i := range aa {
			_, err := db.Exec(ctx, `INSERT INTO media.album_artists (album, artist, artist_type)
				VALUES ($1, $2, $3)`,
				album.MediaID, aa[i].ID, "individual")
			if err != nil {
				errChan <- fmt.Errorf("failed to insert album artist into media.album_artists: %w", err)
			}
		}
		close(errChan)
	}()

	err = <-errChan
	if err != nil {
		return err
	}

	return nil
}

func addTrack(ctx context.Context, db *pgxpool.Pool, track *Track) error {
	_, err := db.Exec(ctx, `
		INSERT INTO media.tracks (media_id, name, duration, lyrics)
		VALUES ($1, $2, $3, $4)`,
		&track.MediaID, &track.Name, &track.Duration, &track.Lyrics)
	if err != nil {
		return fmt.Errorf("failed to insert track into media.tracks: %w", err)
	}

	return nil
}

// PERF: look for areas where concurrency can be added safely here
func (ms *Storage) getAlbum(ctx context.Context, id uuid.UUID) (*Album, error) {
	var album Album

	if err := scn.Get(ctx, ms.db, &album, `SELECT media_id, album_name, release_date, duration
		FROM media.albums 
		WHERE media_id = $1`); err != nil {
		return nil, fmt.Errorf("error getting basic album metadata: %w", err)
	}

	// join to get the name of the artist based on the artist ID, by looking up the person and group tables
	// In person table, we need to select first_name, last_name and nick_names columns
	// then format that using Sprintf
	var albumArtists []AlbumArtist
	if err := scn.Get(ctx, ms.db, &albumArtists, `
SELECT 
    p.first_name, 
    p.last_name, 
    p.nick_names, 
		g.name,
    a.artist, 
    a.artist_type 
FROM 
    media.album_artists AS a
JOIN 
    person AS p ON a.artist = p.id
		group AS g ON a.artist = g.id
WHERE 
    a.album = $1
`, id); err != nil {
		return nil, fmt.Errorf("error querying album artists: %w", err)
	}

	album.AlbumArtists = albumArtists

	var genres []Genre
	if err := scn.Get(ctx, ms.db, &genres, `SELECT genre FROM media.album_genres WHERE album = $1`, id); err != nil {
		return nil, fmt.Errorf("error querying album genres: %w", err)
	}
	album.Genres = genres

	// PERF: find a way to have only one allocation here
	var keywords, keywordsFull []Keyword

	if err := scn.Get(ctx, ms.db, &keywords, `SELECT keyword_id FROM media.album_keywords WHERE album = $1`, id); err != nil {
		return nil, fmt.Errorf("error querying album keywords: %w", err)
	}
	for i := range keywords {
		keyword, err := ms.ks.GetKeywordByID(ctx, keywords[i].ID)
		if err != nil {
			return nil, fmt.Errorf("error getting keyword by id: %w", err)
		}

		keywordsFull = append(keywordsFull, keyword)
	}

	album.Keywords = keywordsFull

	var tracks []Track

	if err := scn.Get(ctx, ms.db, &tracks, `SELECT media_id, name, duration, lyrics
		FROM media.tracks
		WHERE album = $1`, id); err != nil {
		return nil, fmt.Errorf("error querying album tracks: %w", err)
	}

	album.Tracks = tracks

	return &album, nil
}

func (ms *Storage) getTrack(ctx context.Context, id uuid.UUID) (*Track, error) {
	var t *Track
	if err := scn.Select(ctx, ms.db, &t, `SELECT * 
		FROM media.tracks 
		WHERE media_id = $1`); err != nil {
		return nil, fmt.Errorf("error preparing statement: %w", err)
	}

	return t, nil
}

// GetAlbumTracks retrieves the full metadata of given album's tracks based on the album ID
func (ms *Storage) GetAlbumTracks(ctx context.Context, albumID uuid.UUID) ([]Track, error) {
	var tracks []Track
	// Query to fetch tracks and their metadata using a JOIN operation
	query := `
		SELECT t.* FROM media.tracks AS t
		INNER JOIN media.album_tracks AS at ON t.id = at.track_id
		WHERE at.album_id = $1
		ORDER BY at.track_number
	`

	if err := scn.Get(ctx, ms.db, &tracks, query, albumID); err != nil {
		return nil, fmt.Errorf("error querying album tracks: %w", err)
	}
	return tracks, nil
}

func (ms *Storage) GetAlbumTrackIDs(ctx context.Context, albumID uuid.UUID) (dest []*uuid.UUID, err error) {
	query := `
		SELECT track FROM media.album_tracks
		WHERE album = $1
		ORDER BY track_number
	`
	if err := scn.Select(ctx, ms.db, dest, query, albumID); err != nil {
		return nil, fmt.Errorf("error querying album tracks: %w", err)
	}
	return dest, nil
}
