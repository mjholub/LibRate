package media

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"
)

type (
	// nolint:musttag
	Film struct {
		MediaID     *uuid.UUID     `json:"media_id" db:"media_id,pk,unique"`
		Title       string         `json:"title" db:"title" validate:"required"`
		Cast        Cast           `json:"cast"` // this data is stored in the people schema, so no db tag
		ReleaseDate sql.NullTime   `json:"release_date" db:"release_date"`
		Duration    sql.NullTime   `json:"duration" db:"duration"`
		Synopsis    sql.NullString `json:"synopsis" db:"synopsis"`
	}

	TVShow struct {
		MediaID *uuid.UUID `json:"media_id" db:"media_id,pk,unique"`
		Title   string     `json:"title" db:"title" validate:"required"`
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
		Number    uint16        `json:"number" db:"number,autoinc" validate:"required"`
		Title     string        `json:"title" db:"title"`
		Season    uint16        `json:"season" db:"season" validate:"required"`
		Episode   uint16        `json:"episode" db:"episode"`
		AirDate   time.Time     `json:"air_date" db:"air_date"`
		Duration  time.Duration `json:"duration" db:"duration"`
		Languages []string      `json:"languages" db:"languages"`
		Plot      string        `json:"plot" db:"plot"`
	}

	ActorCast struct {
		CastID   int64 `json:"castID" db:"cast_id,pk,unique"`
		PersonID int64 `json:"personID" db:"person_id,pk"`
	}

	DirectorCast struct {
		CastID   int64 `json:"castID" db:"cast_id,pk,unique"`
		PersonID int64 `json:"personID" db:"person_id,pk"`
	}

	Cast struct {
		ID        int64    `json:"ID" db:"cast_id,pk,unique"`
		Actors    []Person `json:"actors" db:"actors"`
		Directors []Person `json:"directors" db:"directors"`
	}
)

func (ms *Storage) getFilm(ctx context.Context, id uuid.UUID) (Film, error) {
	var film Film
	r := ms.db.QueryRow(ctx, "SELECT * FROM media.films WHERE media_id = $1", id)

	if err := r.Scan(&film); err != nil {
		return Film{}, err
	}

	if !film.Synopsis.Valid {
		film.Synopsis.String = "No synopsis available"
	}

	return film, nil
}

func (ms *Storage) getSeries(ctx context.Context, id uuid.UUID) (TVShow, error) {
	var tvshow TVShow
	row := ms.db.QueryRow(ctx, "SELECT * FROM media.tvshows WHERE media_id = $1", id)

	if err := row.Scan(&tvshow); err != nil {
		return TVShow{}, fmt.Errorf("error getting series with ID %s: %w", id.String(), err)
	}

	return tvshow, nil
}

func (f *Film) GetPosterPath(id uuid.UUID) string {
	return "/media/" + id.String() + "/poster.jpg"
}

func (ms *Storage) AddFilm(ctx context.Context, film *Film) error {
	ms.Log.Info().Msg("Adding film \"" + film.Title + "\"")
	// if film has no release date provided yet, set it to 31st December 9999
	// While making the media."media"(created) nullable would seem more intuitive,
	// we do not want this to prevent low quality submissions for stuff that has already been released.
	// This seems like a fair compromise.
	//
	// If an exact release date is not known, there is no way we can prevent users from setting that to 1st January of the
	// actual release year. It should be visible as a tip in the UI, so that if someone reviewing a submission happens to know
	// the exact release date, they can add it.
	if !film.ReleaseDate.Valid {
		film.ReleaseDate.Time = time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
		film.ReleaseDate.Valid = true
	}
	media := Media{
		Title:    film.Title,
		Kind:     "film",
		Created:  film.ReleaseDate.Time,
		Creators: lo.Interleave(film.Cast.Actors, film.Cast.Directors),
	}
	mediaID, err := ms.Add(ctx, &media)
	if err != nil {
		ms.Log.Error().Err(err).Msg("error adding film")
		return err
	}
	tx, err := ms.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	})
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	defer tx.Rollback(ctx)

	ms.Log.Debug().Msgf("Added media with ID " + mediaID.String())
	_, err = tx.Exec(ctx, `
		INSERT INTO media.films (
			media_id, title, cast, release_date, duration, synopsis
		) VALUES (
			:media_id, :title, :release_date, :duration, :synopsis
		)
	`, film)
	if err != nil {
		ms.Log.Error().Err(err).Msg("error adding film")
		return err
	}

	return tx.Commit(ctx)
}

func (ms *Storage) AddCast(ctx context.Context, mediaID uuid.UUID, actors, directors []Person) (castID int64, err error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		if ms.db == nil {
			return 0, fmt.Errorf("no database connection or nil pointer")
		}
		query := `
		INSERT INTO cast VALUES media_id = $1 RETURNING cast_id
		`
		if err = ms.db.QueryRow(ctx, query, mediaID).Scan(&castID); err != nil {
			return 0, fmt.Errorf("error creating cast for media with id %s: %w", mediaID.String(), err)
		}
		// create cast for actors
		// WARN: unsure if the value of castID will be correctly copied in the callback
		// TODO: test this
		lo.ForEach(actors, func(actor Person, _ int) {
			_, err = ms.db.Exec(ctx, `
			INSERT INTO actor_cast (
				cast_id, person_id
			) VALUES (
				$1, $2
			)`, castID, actor.ID)
			if err != nil {
				ms.Log.Error().Err(err).Msgf("error adding actor %s to cast with ID %d", actor.FirstName+actor.LastName, castID)
			}
		})
		// create cast for directors
		lo.ForEach(directors, func(director Person, _ int) {
			_, err = ms.db.Exec(ctx, `
			INSERT INTO director_cast (
				cast_id, person_id
			) VALUES (
				$1, $2
			)`, castID, director.ID)
			if err != nil {
				ms.Log.Error().Err(err).Msgf("error adding director %s to cast with ID %d", director.FirstName+director.LastName, castID)
			}
		})
		return castID, nil
	}
}

func (ms *Storage) GetCast(ctx context.Context, mediaID uuid.UUID) (cast Cast, err error) {
	select {
	case <-ctx.Done():
		return Cast{}, ctx.Err()
	default:
		if ms.db == nil {
			return Cast{}, fmt.Errorf("no database connection or nil pointer")
		}
		// first get the actors ids
		actorIDs := []int64{}
		castQuery := `SELECT id FROM cast WHERE media_id = $1`
		err = ms.db.GetContext(ctx, &cast.ID, castQuery, mediaID)
		if err != nil {
			return Cast{}, fmt.Errorf("error getting cast ID for media with ID %s: %w", mediaID.String(), err)
		}
		query := `SELECT person_id
			FROM cast
			JOIN actor_cast ON actor_cast.cast_id = cast.id
			WHERE cast.media_id = $1`
		err = ms.db.SelectContext(ctx, &actorIDs, query, mediaID)
		if err != nil {
			return Cast{}, fmt.Errorf("error getting cast for media with id %s: %w", mediaID.String(), err)
		}
		var actor Person
		for i := range actorIDs {
			actor, err = ms.Ps.GetPerson(ctx, actorIDs[i])
			if err != nil {
				return Cast{}, fmt.Errorf("error getting actor with id %d: %w", actorIDs[i], err)
			}
			cast.Actors = append(cast.Actors, actor)
		}
		// then get the directors ids
		directorIDs := []int64{}
		query = `SELECT person_id
			FROM cast
			JOIN director_cast ON director_cast.cast_id = cast.id
			WHERE cast.media_id = $1
		`
		err = ms.db.SelectContext(ctx, &directorIDs, query, mediaID)
		if err != nil {
			return Cast{}, fmt.Errorf("error getting cast for media with id %s: %w", mediaID.String(), err)
		}
		for i := range directorIDs {
			director, err := ms.Ps.GetPerson(ctx, directorIDs[i])
			if err != nil {
				return Cast{}, fmt.Errorf("error getting director with id %d: %w", directorIDs[i], err)
			}
			cast.Directors = append(cast.Directors, director)
		}
		return cast, nil
	}
}

func (ms *Storage) UpdateFilm(ctx context.Context, film *Film) error {
	_, err := ms.db.NamedExecContext(ctx, `
		UPDATE media.films SET
			title = :title,
			cast = :cast,
			release_date = :release_date,
			duration = :duration,
			synopsis = :synopsis
		WHERE media_id = :media_id
	`, film)
	if err != nil {
		ms.Log.Error().Err(err).Msg("error updating film")
		return err
	}
	return nil
}
