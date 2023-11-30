package static

import (
	"context"
	"database/sql"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"

	"codeberg.org/mjh/LibRate/db"
	"codeberg.org/mjh/LibRate/internal/lib/thumbnailer"
	"github.com/gofrs/uuid/v5"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type (
	Image struct {
		ID        int64          `json:"id" db:"id,pk,unique,autoinc"`
		Source    string         `json:"source" db:"source"`
		Thumbnail sql.NullString `json:"thumbnail" db:"thumbnail"`
		Alt       sql.NullString `json:"alt" db:"alt"`
	}

	Video struct {
		ID        int64          `json:"id" db:"id,pk,unique,autoinc"`
		Source    string         `json:"source" db:"source"`
		Thumbnail sql.NullString `json:"thumbnail" db:"thumbnail"`
		Alt       sql.NullString `json:"alt" db:"alt"`
	}

	Storage struct {
		db  *sqlx.DB
		Log *zerolog.Logger
	}
)

func NewStorage(db *sqlx.DB, log *zerolog.Logger) *Storage {
	return &Storage{
		db:  db,
		Log: log,
	}
}

func generateThumbnail(source string) (string, error) {
	file, err := os.Open(source)
	if err != nil {
		return "", fmt.Errorf("error generating thumbnail: %w", err)
	}
	defer file.Close()

	thumbProps, err := thumbnailer.Thumbnail(thumbnailer.Dims{Width: 400, Height: 400}, source)
	if err != nil {
		return "", fmt.Errorf("error generating thumbnail: %w", err)
	}

	thumb, err := saveThumbToFile(&thumbProps, source)
	if err != nil {
		return "", fmt.Errorf("error generating thumbnail: %w", err)
	}
	return thumb, nil
}

// saveThumbToFile encodes the thumbnail image properties obtained using the thumbnailer
// TODO: use mo.Result to simplify error handling when this func is called?
func saveThumbToFile(thumb *image.Image, outPath string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error saving thumbnail: %w", err)
	}
	thumbFile, err := os.Create(filepath.Join(cwd, "static", outPath, "thumbnail.jpg"))
	if err != nil {
		return "", fmt.Errorf("error saving thumbnail: %w", err)
	}
	defer thumbFile.Close()

	err = jpeg.Encode(thumbFile, *thumb, nil)
	if err != nil {
		return "", fmt.Errorf("error saving thumbnail: %w", err)
	}

	return outPath, nil
}

func (s *Storage) AddVideo(v *Video) error {
	thumb, err := generateThumbnail(v.Source)
	if err != nil {
		return fmt.Errorf("error adding video: %w", err)
	}
	s.Log.Info().Msgf("Generated thumbnail for video %s, \nPath: %s", v.Source, thumb)

	_, err = s.db.ExecContext(context.Background(), `INSERT INTO cdn.videos (source, thumbnail, alt)
		VALUES ($1, $2, $3)`, v.Source, thumb, v.Alt)
	if err != nil {
		return fmt.Errorf("error adding video: %w", err)
	}

	return nil
}

func (s *Storage) GetImageSource(ctx context.Context, id int64) (source string, err error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		stmt, err := s.db.PreparexContext(ctx, `SELECT source FROM cdn.images WHERE id = $1`)
		if err != nil {
			return "", fmt.Errorf("error retrieving image source: %w", err)
		}
		defer stmt.Close()

		err = stmt.GetContext(ctx, &source, id)
		if err != nil {
			return "", fmt.Errorf("error retrieving image source: %w", err)
		}
		return source, nil
	}
}

func (s *Storage) AddImage(ctx context.Context, imageType, ext, uploader string, mediaID *uuid.UUID) (dest string, id int64, err error) {
	select {
	case <-ctx.Done():
		return "", 0, ctx.Err()
	default:
		stmt, err := s.db.PreparexContext(ctx, `INSERT INTO cdn.images (source, uploader) VALUES ($1, $2) RETURNING id`)
		if err != nil {
			return "", 0, fmt.Errorf("error adding file: %w", err)
		}
		defer stmt.Close()

		// TODO: test for path traversal
		uploader = db.Sanitize([]string{uploader})[0]

		switch imageType {
		case "profile":
			dest = fmt.Sprintf("static/img/profile/%s.%s", uploader, ext)
		case "album_cover":
			if mediaID == nil {
				return "", 0, fmt.Errorf("error adding file: %w", err)
			}
			if mediaID.String() == "" {
				return "", 0, fmt.Errorf("error adding file: %w", err)
			}
			dest = fmt.Sprintf("static/img/music/%s.%s", mediaID.String(), ext)
		default:
			return "", 0, fmt.Errorf("unknown image type %s", imageType)
		}

		err = stmt.GetContext(ctx, &id, dest, uploader)
		if err != nil {
			return "", 0, fmt.Errorf("error adding file: %w", err)
		}
		return dest, id, nil
	}
}

func (s *Storage) GetOwner(ctx context.Context, imageID int64) (owner string, err error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		// not using prepared statements because the query parameter is numeric
		err = s.db.GetContext(ctx, &owner, `SELECT uploader FROM cdn.images WHERE id = $1`, imageID)
		if err != nil {
			return "", fmt.Errorf("error retrieving image owner: %w", err)
		}
		return owner, nil
	}
}

// DeleteImage looks up the path of the image to delete based on it's id, then deletes the database record and returns the path to be deleted by the controller
func (s *Storage) DeleteImage(ctx context.Context, imageID int64) (path string, err error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return "", fmt.Errorf("error deleting image: %w", err)
		}
		defer tx.Rollback()
		err = s.db.GetContext(ctx, &path, `SELECT source FROM cdn.images WHERE id = $1`, imageID)
		if err != nil {
			return "", fmt.Errorf("error retrieving image path for deletion: %w", err)
		}
		err = s.db.GetContext(ctx, &path, `DELETE FROM cdn.images WHERE id = $1`, imageID)
		if err != nil {
			return "", fmt.Errorf("error deleting image: %w", err)
		}
		err = tx.Commit()
		if err != nil {
			return "", fmt.Errorf("error deleting image: %w", err)
		}
		return path, nil
	}
}
