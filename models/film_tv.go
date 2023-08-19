package models

import (
	"context"
	"database/sql"
	"time"

	"github.com/gofrs/uuid/v5"
)

type (
	Film struct {
		MediaID     *uuid.UUID   `json:"media_id" db:"media_id,pk,unique"`
		Title       string       `json:"title" db:"title"`
		Cast        Cast         `json:"cast" db:"cast"`
		ReleaseDate sql.NullTime `json:"release_date" db:"release_date"`
	}

	TVShow struct {
		MediaID *uuid.UUID `json:"media_id" db:"media_id,pk,unique"`
		Title   string     `json:"title" db:"title"`
		Cast    Cast       `json:"cast" db:"cast"`
		Year    int        `json:"year" db:"year"`
		Active  bool       `json:"active" db:"active"`
		Seasons []Season   `json:"seasons" db:"seasons"`
		Studio  Studio     `json:"studio" db:"studio"`
	}

	Season struct {
		MediaID  *uuid.UUID `json:"media_id" db:"media_id,pk,unique"`
		ShowID   *uuid.UUID `json:"show_id" db:"show_id,pk,unique"`
		Number   uint8      `json:"number" db:"number"`
		Episodes []Episode  `json:"episodes" db:"episodes"`
	}

	Episode struct {
		MediaID   *uuid.UUID    `json:"media_id" db:"media_id,pk,unique"`
		ShowID    *uuid.UUID    `json:"show_id" db:"show_id,pk,unique"`
		SeasonID  *uuid.UUID    `json:"season_id" db:"season_id,pk,unique"`
		Number    uint16        `json:"number" db:"number,autoinc"`
		Title     string        `json:"title" db:"title"`
		Season    uint16        `json:"season" db:"season"`
		Episode   uint16        `json:"episode" db:"episode"`
		AirDate   time.Time     `json:"air_date" db:"air_date"`
		Duration  time.Duration `json:"duration" db:"duration"`
		Languages []string      `json:"languages" db:"languages"`
		Plot      string        `json:"plot" db:"plot"`
	}

	// TODO: add more fields
	Cast struct {
		Actors    []Person `json:"actors" db:"actors"`
		Directors []Person `json:"directors" db:"directors"`
	}
)

func (ms *MediaStorage) getFilm(ctx context.Context, id uuid.UUID) (Film, error) {
	var film Film
	err := ms.db.GetContext(ctx, &film, "SELECT * FROM media.films WHERE media_id = ?", id)
	if err != nil {
		return Film{}, err
	}

	return film, nil
}

func (ms *MediaStorage) getSeries(ctx context.Context, id uuid.UUID) (TVShow, error) {
	var tvshow TVShow
	err := ms.db.GetContext(ctx, &tvshow, "SELECT * FROM media.tvshows WHERE media_id = ?", id)
	if err != nil {
		return TVShow{}, err
	}

	return tvshow, nil
}

func (f Film) IsMedia() bool {
	return true
}

func (ts TVShow) IsMedia() bool {
	return true
}

func (s Season) IsMedia() bool {
	return true
}

func (e Episode) IsMedia() bool {
	return true
}
