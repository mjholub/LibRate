package models

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jmoiron/sqlx"
	"github.com/samber/mo"
)

type (
	Album struct {
		MediaID      *uuid.UUID                   `json:"media_id" db:"media_id,pk,unique"`
		Name         string                       `json:"name" db:"name"`
		AlbumArtists mo.Either[[]Person, []Group] `json:"album_artists" db:"album_artists"`
		ReleaseDate  time.Time                    `json:"release_date" db:"release_date"`
		Genres       []Genre                      `json:"genres,omitempty" db:"genres"`
		//	Studio       Studio                       `json:"studio,omitempty" db:"studio"`
		Keywords []Keyword `json:"keywords,omitempty" db:"keywords"`
		Duration time.Time `json:"duration" db:"duration"`
		Tracks   []Track   `json:"tracks" db:"tracks"`
		//	Languages int16         `json:"languages" db:"languages,omitempty"`
	}

	Track struct {
		MediaID   *uuid.UUID                   `json:"media_id" db:"media_id,pk,unique"`
		Name      string                       `json:"name" db:"name"`
		Artists   mo.Either[[]Person, []Group] `json:"artists" db:"artists"`
		Duration  time.Duration                `json:"duration" db:"duration"`
		Lyrics    string                       `json:"lyrics,omitempty" db:"lyrics"`
		Languages []string                     `json:"languages,omitempty" db:"languages"`
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
		_, err := db.ExecContext(ctx, "INSERT INTO media.album_genres (album, genre) VALUES ($1, $2)",
			album.MediaID, album.Genres[i].ID)
		if err != nil {
			return fmt.Errorf("failed to insert album genre into media.album_genres: %w", err)
		}
	}

	errChan := make(chan error)
	album.AlbumArtists.ForEach(func(artists []Person) {
		for i := range artists {
			_, err = db.ExecContext(ctx, "INSERT INTO media.album_artists (album, person_artist) VALUES ($1, $2)",
				album.MediaID, artists[i].ID)
			if err != nil {
				errChan <- fmt.Errorf("failed to insert album artist into media.album_artists: %w", err)
			}
		}
	}, func(groups []Group) {
		for i := range groups {
			_, err = db.ExecContext(ctx, "INSERT INTO media.album_artists (album, group_artist) VALUES ($1, $2)",
				album.MediaID, groups[i].ID)
			if err != nil {
				errChan <- fmt.Errorf("failed to insert album artist into media.album_artists: %w", err)
			}
		}
	})
	err = <-errChan
	if err != nil {
		close(errChan)
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

	artists, _ := track.Artists.Left()
	for i := range artists {
		_, err := db.ExecContext(ctx, "INSERT INTO media.track_artists (track, artist) VALUES ($1, $2)",
			&track.MediaID, artists[i].ID)
		if err != nil {
			return fmt.Errorf("failed to insert track artist into media.track_artists: %w", err)
		}
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

	rows, err := ms.db.QueryContext(ctx, `SELECT person_artist, group_artist
		FROM media.album_artists
		WHERE album = $1`, id)
	if err != nil {
		return Album{}, fmt.Errorf("error querying album artists: %w", err)
	}
	defer rows.Close()
	var (
		personArtists []Person
		groupArtists  []Group
		artists       []mo.Either[Person, Group]
	)

	for _, artist := range artists {
		if artist.IsLeft() {
			person, _ := artist.Left()
			personArtists = append(personArtists, person)
		} else {
			group, _ := artist.Right()
			groupArtists = append(groupArtists, group)
		}
	}

	if len(personArtists) > 0 {
		album.AlbumArtists = mo.Left[[]Person, []Group](personArtists)
	} else {
		album.AlbumArtists = mo.Right[[]Person, []Group](groupArtists)
	}

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

	return album, nil
}

func (ms *MediaStorage) getTrack(ctx context.Context, id uuid.UUID) (Track, error) {
	stmt, err := ms.db.PrepareContext(ctx, "SELECT * FROM tracks WHERE media_id = $1")
	if err != nil {
		return Track{}, fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, id)
	var track Track
	if err := row.Scan(&track); err != nil {
		return Track{}, fmt.Errorf("error scanning row: %w", err)
	}

	return track, nil
}
