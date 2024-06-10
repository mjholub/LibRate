package media

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	dblib "codeberg.org/mjh/LibRate/db"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type (
	Storer[T any] interface {
		Get(ctx context.Context, key string) (T, error)
		GetAll() ([]T, error)
		Add(ctx context.Context, db *pgxpool.Pool, props Media) (uuid.UUID, error)
		Update(ctx context.Context, key string, value T) error
		Delete(ctx context.Context, key string) error
	}

	// nolint:musttag // false positive, can only annotate fields, not types
	Media struct {
		ID       uuid.UUID     `json:"id" db:"id,pk,unique"`
		Title    string        `json:"title" db:"title"`
		Kind     string        `json:"kind" db:"kind"`
		Created  time.Time     `json:"keywords,omitempty" db:"created"`
		Creator  sql.NullInt32 `json:"creator,omitempty" db:"creator"`
		Creators []Person      `json:"creators,omitempty"` // no db tag, we're using a junction table
		Added    time.Time     `json:"added,omitempty" db:"added"`
		Modified sql.NullTime  `json:"modified,omitempty" db:"modified"`
	}

	// used in search
	SimplifiedMedia struct {
		Title       string `json:"title" db:"title"`
		Kind        string `json:"kind" db:"kind"`
		ImageSource string `json:"image_source" db:"source"`
	}

	Details struct {
		Kind string      `json:"kind" db:"kind"`
		Data interface{} `json:"details" db:"details"`
	}

	//nolint:revive // renaming this to Object would be confusing
	MediaObject interface {
		Book | Album | Track | TVShow | Season | Episode
	}

	GenresOrGenreNames interface {
		[]Genre | []string
	}

	// Genre does not have a UUID due to parent-child relationships
	Genre struct {
		ID          int64              `json:"id" db:"id,pk,autoinc"`
		Kinds       pq.StringArray     `json:"kind" db:"kind" enum:"music,film,tv,book,game" example:"music"`
		Name        string             `json:"name" db:"name" example:"Black Metal"`
		Description []GenreDescription `json:"description,omitempty" db:"-"`
		//	DescLong    string   `json:"desc_long" db:"desc_long"`
		Characteristics []string `json:"keywords" db:"-" example:"['dark', 'gloomy', 'atmospheric', 'raw', 'underproduced']"`
		ParentGenreID   *int64   `json:"parent_genre,omitempty" db:"parent,omitempty"`
		Children        []int64  `json:"children,omitempty" db:"children,omitempty"`
	}

	GenreCharacteristics struct {
		ID         int64          `json:"id" db:"id,pk,autoinc"`
		Name       string         `json:"name" db:"name"`
		Descripion sql.NullString `json:"description,omitempty" db:"description"`
	}

	GenreDescription struct {
		GenreID     int64  `json:"genre_id" db:"genre_id" example:"2958"`
		Language    string `json:"language" db:"language" example:"en"`
		Description string `json:"description" db:"description" example:"Typified by highly distorted, trebly, tremolo-picked guitars, blast beats, double kick drumming, shrieked vocals, and raw, underproduced sound that often favors atmosphere over technical skills and melody."`
	}

	Storage struct {
		db  *pgxpool.Pool
		Log *zerolog.Logger
		ks  *KeywordStorage
		Ps  *PeopleStorage
	}
)

func NewStorage(db *pgxpool.Pool, l *zerolog.Logger) *Storage {
	ks := NewKeywordStorage(db, l)
	Ps := NewPeopleStorage(db, l)
	return &Storage{db: db, Log: l, ks: ks, Ps: Ps}
}

// Get scans into a complete Media struct
// In most cases though, all we need is an intermediate, partial instance with the UUID and Kind fields
// to be passed to GetDetails
func (ms *Storage) Get(ctx context.Context, id uuid.UUID) (media *Media, err error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		result, err := dblib.SerializableParametrizedTx[*Media](ctx, ms.db, "get_media",
			`SELECT * FROM media.media WHERE id = $1`,
			map[string]string{"media_id": id.String()},
			id)
		if err != nil {
			return nil, fmt.Errorf("error querying media: %v", err)
		}
		return result[0], nil
	}
}

func (ms *Storage) GetImagePath(ctx context.Context, id uuid.UUID) (path string, err error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		result, err := dblib.SerializableParametrizedTx[string](ctx, ms.db, "get_image_path",
			`SELECT i.source FROM media.media_images AS mi
		JOIN cdn.images AS i ON mi.image_id = i.id
		WHERE mi.media_id = $1
		LIMIT 1`,
			map[string]string{"media_id": id.String()},
			id)
		if err != nil {
			return "", fmt.Errorf("error querying image path: %v", err)
		}

		return result[0], nil
	}
}

// TODO: a method to get thumbnail path(s) (?)

func (ms *Storage) GetKind(ctx context.Context, id uuid.UUID) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		result, err := dblib.SerializableParametrizedTx[string](ctx, ms.db, "get_kind",
			`SELECT kind FROM media.media WHERE id = $1`,
			map[string]string{"media_id": id.String()},
			id)
		if err != nil {
			return "", fmt.Errorf("error querying kind: %v", err)
		}

		return result[0], nil
	}
}

// GetGenres returns all genres for specified media type.
// parameter all specifies whether to return all genres or only top-level ones.
// variadic argument columns specifies which columns to return.
// In HTTP layer, columns are specified either in the JSON request body (as an array of strings)
// The name column can also be accessed with `names_only` boolean query parameter.
// Fetching of all genres is specified by the `all` query parameter (which does not require a value).
func GetGenres[G GenresOrGenreNames](
	ms *Storage,
	// nolint:revive // hacky generic dependency injection so would-be receiver should be the 1st arg
	ctx context.Context,
	kind string,
	all bool,
	columns ...string,
) ([]G, error) {
	if len(columns) > 0 {
		validColumns := []string{"id", "kinds", "name", "parent", "children"}
		for i := range columns {
			if !lo.Contains(validColumns, columns[i]) {
				return nil, fmt.Errorf("invalid column name: %v", columns[i])
			}
		}
		ms.Log.Debug().Msg("validated columns")
	}
	const baseGenres = "WHERE $1 = ANY(kinds) AND parent IS NULL"
	const allGenres = "WHERE $1 = ANY(kinds)"

	queryTemplate := "SELECT %v FROM media.genres %v"
	whereClause := baseGenres

	if all {
		ms.Log.Debug().Msg("fetching all genres")
		whereClause = allGenres
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		stmt := fmt.Sprintf(queryTemplate, strings.Join(columns, ", "), whereClause)

		genreList, err := dblib.SerializableParametrizedTx[G](ctx, ms.db, "get_genres",
			stmt, map[string]any{
				"kind":    kind,
				"all":     all,
				"columns": columns,
			},
			kind)

		if err != nil {
			return nil, fmt.Errorf("error preparing statement: %v", err)
		}

		return genreList, nil
	}
}

func (ms *Storage) GetGenre(ctx context.Context, kind, lang, name string) (genre *Genre, err error) {
	title := cases.Title(language.AmericanEnglish)
	name = title.String(strings.ReplaceAll(name, "_", " "))
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		genre := Genre{
			Kinds: pq.StringArray{kind},
		}
		rows, err := ms.db.Query(ctx, `
			SELECT id, name, parent, children FROM media.genres WHERE $1 = ANY(kinds) AND name = $2`,
			kind, name)
		if err != nil {
			return nil, fmt.Errorf("error querying genre rows: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			if err = pgxscan.ScanRow(&genre, rows); err != nil {
				return nil, fmt.Errorf("error scanning row: %v", err)
			}
		}
		dc, err := ms.db.Acquire(ctx)
		if err != nil {
			return nil, fmt.Errorf("error acquiring connection: %v", err)
		}
		defer dc.Release()

		var description string
		err = dc.QueryRow(ctx, `
			SELECT description FROM media.genre_descriptions	
			WHERE genre_id = (SELECT id FROM media.genres WHERE $1 = ANY(kinds) AND name = $2) 
			AND language = $3
					`, kind, name, lang).Scan(&description)
		if err != nil {
			return nil, fmt.Errorf("error querying rows for description: %v", err)
		}

		genre.Description = []GenreDescription{
			{
				GenreID:     genre.ID,
				Language:    lang,
				Description: description,
			},
		}

		ms.Log.Debug().Msgf("genre: %v", genre)

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating rows: %v", err)
		}
		return &genre, nil
	}
}

func (ms *Storage) GetDetails(
	ctx context.Context,
	mediaKind string,
	id uuid.UUID,
) (interface{}, error) {
	switch mediaKind {
	case "book":
		return ms.getBook(ctx, id)
	case "album":
		return ms.getAlbum(ctx, id)
	case "track":
		return ms.getTrack(ctx, id)
	case "film":
		return ms.getFilm(ctx, id)
	case "tv_show":
		return ms.getSeries(ctx, id)
	default:
		return nil, fmt.Errorf("unknown media kind")
	}
}

// mwks - media IDs with their corresponding kind
func (ms *Storage) GetRandom(ctx context.Context, count int, blacklistKinds ...string) (
	mwks map[uuid.UUID]string, err error,
) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		result, err := dblib.SerializableParametrizedTx[map[uuid.UUID]string](ctx, ms.db, "get_random_media",
			`SELECT id, kind
			FROM media.media 
			WHERE kind != ALL($1)
			ORDER BY RANDOM()
			LIMIT $2`,
			map[string]any{"blacklist": blacklistKinds, "count": count},
			blacklistKinds, count)

		if err != nil {
			return nil, fmt.Errorf("error querying random media: %v", err)
		}

		return result[0], nil
	}
}

// Add is a generic method that adds an object to the media.media table. It needs to be run
// BEFORE the object is added to its respective table, since it needs the media ID to be
// generated first.
func (ms *Storage) Add(ctx context.Context, props *Media) (mediaID *uuid.UUID, err error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// early return on faulty db connection
		if ms.db == nil {
			return nil, fmt.Errorf("no database connection or nil pointer")
		}

		stmt, err := dblib.SerializableParametrizedTx[*uuid.UUID](ctx,
			ms.db,
			"add_media",
			`	
		INSERT INTO media.media (
			title, kind, created
		) VALUES (
			$1, $2, $3
		)
		RETURNING id
		`,
			map[string]any{
				"title":   props.Title,
				"kind":    props.Kind,
				"created": props.Created,
			},
			props.Title, props.Kind, props.Created)
		if err != nil {
			return nil, fmt.Errorf("error preparing statement: %v", err)
		}

		mediaID = stmt[0]
		if mediaID == nil {
			return nil, fmt.Errorf("no media ID returned")
		}

		err = ms.AddCreators(ctx, *mediaID, props.Creators)
		// handle the case in which the said person is not in the database
		if err == sql.ErrNoRows {
			ms.Log.Warn().Msg("no rows were affected")
			return nil, fmt.Errorf("no rows were affected: %v", err)
		}
		if err != nil {
			return nil, fmt.Errorf("error adding creators: %v", err)
		}
		return mediaID, nil
	}
}

func (ms *Storage) AddCreators(ctx context.Context, id uuid.UUID, creators []Person) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// WARN: cannot use dblib.SerialParametrizedUnaryTx,
		// because Go type system fucking sucks and I regret picking it
		// to build LibRate in the first place
		// (i.e. adding it to this stupid convertParam type switch
		// would cause a cyclic dependency).
		// Of course we can always use mapstructure, but honestly,
		// how many annotations are we going to put in our structs?
		// Soon I'll either have to disable line length limit in linter or
		// make my struct definitions unreadable.
		// jfc
		// https://super8.absturztau.be/watch?v=aSEQfqNYNAc
		tx, err := ms.db.BeginTx(ctx, pgx.TxOptions{
			IsoLevel: pgx.Serializable,
		})
		const (
			q   = "add-creators"
			sql = `INSERT INTO media.media_creators (
			media_id, creator_id
		) VALUES (
			$1, $2)`
		)
		if err != nil {
			return dblib.TxErr(q, map[string]any{
				"media_id": id.String(),
				"creators": creators,
			}, err)
		}
		defer tx.Rollback(ctx)

		_, err = tx.Prepare(ctx, q, sql)
		if err != nil {
			return fmt.Errorf("error preparing statement: %w", err)
		}

		if err = tx.Commit(ctx); err != nil {
			return fmt.Errorf("error executing statement: %w", err)
		}
		return nil
	}
}

func (ms *Storage) GetAll() ([]*interface{}, error) {
	return nil, nil
}

func (ms *Storage) Update(ctx context.Context, key string, value interface{}, mediaID uuid.UUID) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// early return on faulty db connection
		if ms.db == nil {
			return fmt.Errorf("no database connection or nil pointer")
		}
		_, err := ms.db.Exec(ctx, `
		UPDATE media.media
		SET $1 = $2
		WHERE id = $3
		`, key, value, mediaID)
		if err != nil {
			return fmt.Errorf("error updating media: %v", err)
		}
		return nil
	}
}

func (ms *Storage) Delete(ctx context.Context, mediaID uuid.UUID) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if err := dblib.SerialParametrizedUnaryTx(ctx, ms.db, "delete_media",
			`DELETE FROM media.media WHERE id = $1`,
			map[string]string{"media_id": mediaID.String()},
			mediaID,
		); err != nil {
			return fmt.Errorf("error deleting media with ID %s: %w", mediaID.String(), err)
		}
		return nil
	}
}
