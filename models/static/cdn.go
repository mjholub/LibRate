package static

import (
	"context"
	"database/sql"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"

	"codeberg.org/mjh/LibRate/internal/lib/thumbnailer"
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
