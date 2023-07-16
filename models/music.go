package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	//"github.com/samber/mo"
)

type (
	Album struct {
		MediaID      *uuid.UUID     `json:"media_id" db:"media_id,pk,unique"`
		Name         string         `json:"name" db:"name"`
		AlbumArtists AlbumArtist    `json:"album_artists" db:"album_artists"`
		ImagePaths   pq.StringArray `json:"image_paths,omitempty"` // we make use of a junction table that utilizes the image IDs
		ReleaseDate  time.Time      `json:"release_date" db:"release_date"`
		Genres       []Genre        `json:"genres,omitempty" db:"genres"`
		//	Studio       Studio                       `json:"studio,omitempty" db:"studio"`
		Keywords []Keyword    `json:"keywords,omitempty" db:"keywords"`
		Duration sql.NullTime `json:"duration,omitempty" db:"duration"`
		Tracks   []Track      `json:"tracks,omitempty" db:"tracks"`
		//	Languages int16         `json:"languages" db:"languages,omitempty"`
	}

	AlbumArtist struct {
		PersonArtists []Person `json:"person_artist,omitempty" db:"person_artist"`
		GroupArtists  []Group  `json:"group_artist,omitempty" db:"group_artist"`
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

	MusicValues interface {
		string | []string | time.Duration | []time.Duration | []uuid.UUID | []Person | []Group | []Genre | []Studio | []Track | time.Time | uuid.UUID
	}
)

func addAlbum(ctx context.Context, db *sqlx.DB, album Album) error {
	// Insert the album into the media.albums table
	_, err := db.ExecContext(ctx, `
		INSERT INTO media.albums (media_id, name, release_date, duration)
		VALUES ($1, $2, $3, $4)`,
		album.MediaID, album.Name, album.ReleaseDate, album.Duration)
	if err != nil {
		return fmt.Errorf("failed to insert album into media.albums: %w", err)
	}
	// Insert keywords into the junction table
	_, err = db.ExecContext(ctx, `
		INSERT INTO media.album_keywords (album, keyword)
		VALUES ($1, $2)`,
		album.MediaID, album.Keywords)
	if err != nil {
		return fmt.Errorf("failed to insert album keywords into media.album_keywords: %w", err)
	}

	// Insert the genres into the media.album_genres table
	for i := range album.Genres {
		_, err = db.ExecContext(ctx, "INSERT INTO media.album_genres (album, genre) VALUES ($1, $2)",
			album.MediaID, album.Genres[i].ID)
		if err != nil {
			return fmt.Errorf("failed to insert album genre into media.album_genres: %w", err)
		}
	}

	errChan := make(chan error)
	go func() {
		for _, artist := range album.AlbumArtists.PersonArtists {
			_, err = db.ExecContext(ctx, "INSERT INTO media.album_artists (album, person_artist) VALUES ($1, $2)",
				album.MediaID, artist.ID)
			if err != nil {
				errChan <- fmt.Errorf("failed to insert album artist into media.album_artists: %w", err)
				return
			}
		}
		for _, group := range album.AlbumArtists.GroupArtists {
			_, err = db.ExecContext(ctx, "INSERT INTO media.album_artists (album, group_artist) VALUES ($1, $2)",
				album.MediaID, group.ID)
			if err != nil {
				errChan <- fmt.Errorf("failed to insert album artist into media.album_artists: %w", err)
				return
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

func addTrack(ctx context.Context, db *sqlx.DB, track *Track) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO media.tracks (media_id, name, duration, lyrics)
		VALUES ($1, $2, $3, $4)`,
		&track.MediaID, &track.Name, &track.Duration, &track.Lyrics)
	if err != nil {
		return fmt.Errorf("failed to insert track into media.tracks: %w", err)
	}

	return nil
}

func (ms *MediaStorage) getAlbum(ctx context.Context, id uuid.UUID) (Album, error) {
	stmt, err := ms.db.PrepareContext(ctx, `SELECT media_id, album_name, release_date, duration
		FROM media.albums 
		WHERE media_id = $1`)
	if err != nil {
		return Album{}, fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, id)
	var album Album
	err = row.Scan(&album.MediaID, &album.Name, &album.ReleaseDate, &album.Duration)
	if err != nil {
		return Album{}, fmt.Errorf("error scanning row: %w", err)
	}
	rows, err := ms.db.QueryContext(ctx, "SELECT person_artist FROM media.album_artists WHERE album = $1", album.MediaID)
	if err != nil {
		return Album{}, fmt.Errorf("error querying album artists: %w", err)
	}

	for rows.Next() {
		var personID int32
		err = rows.Scan(&personID)
		if err != nil {
			return Album{}, err
		}
		var person Person
		person, err = ms.ps.GetPerson(ctx, personID)
		if err != nil {
			return Album{}, fmt.Errorf("error getting person names: %w", err)
		}
		album.AlbumArtists.PersonArtists = append(album.AlbumArtists.PersonArtists, person)
	}

	// Fetch group artists
	rows, err = ms.db.QueryContext(ctx, "SELECT group_artist FROM media.album_artists WHERE album = $1", album.MediaID)
	if err != nil {
		return Album{}, fmt.Errorf("error querying album artists: %w", err)
	}

	for rows.Next() {
		var groupID int32
		err = rows.Scan(&groupID)
		if err != nil {
			return Album{}, fmt.Errorf("error scanning row: %w", err)
		}
		var group Group
		group, err = ms.ps.GetGroup(ctx, groupID)
		if err != nil {
			return Album{}, fmt.Errorf("error getting group name: %w", err)
		}
		album.AlbumArtists.GroupArtists = append(album.AlbumArtists.GroupArtists, group)
	}
	ms.Log.Info().Msgf("person artists: %v", album.AlbumArtists.PersonArtists)
	ms.Log.Info().Msgf("group artists: %v", album.AlbumArtists.GroupArtists)

	rows, err = ms.db.QueryContext(ctx, `SELECT genre FROM media.album_genres WHERE album = $1`, id)
	if err != nil {
		return Album{}, fmt.Errorf("error querying album genres: %w", err)
	}
	defer rows.Close()
	var genres []Genre
	for rows.Next() {
		var genre Genre
		err = rows.Scan(&genre.ID)
		if err != nil {
			return Album{}, fmt.Errorf("error scanning row: %w", err)
		}
		genres = append(genres, genre)
	}

	album.Genres = genres

	rows, err = ms.db.QueryContext(ctx, `SELECT keyword_id FROM media.album_keywords WHERE album = $1`, id)
	if err != nil {
		return Album{}, fmt.Errorf("error querying album keywords: %w", err)
	}
	defer rows.Close()
	var keywords []Keyword
	var keyword Keyword
	for rows.Next() {
		var keywordID int32
		err = rows.Scan(&keywordID)
		if err != nil {
			return Album{}, fmt.Errorf("error scanning row: %w", err)
		}
		keyword, err = ms.ks.GetKeywordByID(ctx, keywordID)
		if err != nil {
			return Album{}, fmt.Errorf("error getting keyword by id: %w", err)
		}

		keywords = append(keywords, keyword)
	}

	album.Keywords = keywords

	rows, err = ms.db.QueryContext(ctx, `SELECT media_id, name, duration, lyrics
		FROM media.tracks
		WHERE album = $1`, id)
	if err != nil {
		return Album{}, fmt.Errorf("error querying album tracks: %w", err)
	}
	defer rows.Close()
	var tracks []Track
	for rows.Next() {
		var track Track
		err = rows.Scan(&track.MediaID, &track.Name, &track.Duration, &track.Lyrics)
		if err != nil {
			return Album{}, fmt.Errorf("error scanning row: %w", err)
		}
		tracks = append(tracks, track)
	}

	album.Tracks = tracks
	ms.Log.Trace().Msgf("album: %v", album)

	return album, nil
}

func (ms *MediaStorage) getTrack(ctx context.Context, id uuid.UUID) (Track, error) {
	stmt, err := ms.db.PreparexContext(ctx, `SELECT * 
		FROM media.tracks 
		WHERE media_id = $1`)
	if err != nil {
		return Track{}, fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	row := stmt.QueryRowxContext(ctx, id)
	var track Track
	if err := row.StructScan(&track); err != nil {
		return Track{}, fmt.Errorf("error scanning row: %w", err)
	}

	return track, nil
}
