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
		Studio       Studio                       `json:"studio,omitempty" db:"studio"`
		Keywords     []string                     `json:"keywords,omitempty" db:"keywords"`
		Duration     time.Duration                `json:"duration" db:"duration"`
		Tracks       []Track                      `json:"tracks" db:"tracks"`
		Languages    []string                     `json:"languages" db:"languages,omitempty"`
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
		INSERT INTO media.albums (media_id, name, release_date, keywords, duration)
		VALUES (?, ?, ?, ?, ?)`,
		album.MediaID, album.Name, album.ReleaseDate, album.Keywords, album.Duration)
	if err != nil {
		return err
	}

	// Insert the genres into the media.album_genres table
	for i := range album.Genres {
		_, err := db.ExecContext(ctx, "INSERT INTO media.album_genres (album, genre) VALUES (?, ?)", album.MediaID, album.Genres[i].ID)
		if err != nil {
			return fmt.Errorf("failed to insert album genre into media.album_genres: %w", err)
		}
	}

	return nil
}

func addTrack(ctx context.Context, db *sqlx.DB, track Track) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO media.tracks (media_id, name, duration, lyrics)
		VALUES (?, ?, ?, ?)`,
		track.MediaID, track.Name, track.Duration, track.Lyrics)
	if err != nil {
		return fmt.Errorf("failed to insert track into media.tracks: %w", err)
	}

	artists, _ := track.Artists.Left()
	for i := range artists {
		_, err := db.ExecContext(ctx, "INSERT INTO media.track_artists (track, artist) VALUES (?, ?)", track.MediaID, artists[i].ID)
		if err != nil {
			return fmt.Errorf("failed to insert track artist into media.track_artists: %w", err)
		}
	}

	return nil
}
