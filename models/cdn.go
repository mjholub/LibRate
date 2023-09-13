package models

import (
	"context"
	"database/sql"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"

	// WARN: using a fork due to a trivial CGO bug in the original package
	// FIXME: this package has been abandoned. If possible try finding an alternative or writing one, preferably in pure Go
	thumbnailer "github.com/abdelrahmanmostafa21/go-thumbnailer/v2"
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

	StaticStorage struct {
		db  *sqlx.DB
		Log *zerolog.Logger
	}
)

// TODO: move below the following func
func getVideoDimensions(vPath string) (thumbnailer.Dims, error) {
	f, err := os.Open(vPath)
	if err != nil {
		return thumbnailer.Dims{}, fmt.Errorf("failed to determine video dimensions: %w", err)
	}
	defer f.Close()

	ctx, err := thumbnailer.NewFFContext(f)
	if err != nil {
		return thumbnailer.Dims{}, fmt.Errorf("failed to determine video dimensions: %w", err)
	}
	defer ctx.Close()

	dims, err := ctx.Dims()
	if err != nil {
		return thumbnailer.Dims{}, fmt.Errorf("failed to determine video dimensions: %w", err)
	}

	return dims, nil
}

func generateThumbnail(source string) (string, error) {
	file, err := os.Open(source)
	if err != nil {
		return "", fmt.Errorf("error generating thumbnail: %w", err)
	}
	defer file.Close()

	originalDims, err := getVideoDimensions(source)
	if err != nil {
		return "", fmt.Errorf("error generating thumbnail: %w", err)
	}
	totalPixels := originalDims.Width * originalDims.Height

	switch {
	case totalPixels <= 640*480 && totalPixels > 320*240:
		_, thumbProps, err := thumbnailer.Process(file, thumbnailer.Options{
			ThumbDims: thumbnailer.Dims{
				Width:  originalDims.Width / 2,
				Height: originalDims.Height / 2,
			},
		})
		if err != nil {
			return "", fmt.Errorf("error generating thumbnail: %w", err)
		}
		thumb, err := saveThumbToFile(&thumbProps, source)
		if err != nil {
			return "", fmt.Errorf("error generating thumbnail: %w", err)
		}
		return thumb, nil
	case totalPixels <= 1280*720:
		_, thumbProps, err := thumbnailer.Process(file, thumbnailer.Options{
			ThumbDims: thumbnailer.Dims{
				Width:  originalDims.Width / 4,
				Height: originalDims.Height / 4,
			},
		})
		if err != nil {
			return "", fmt.Errorf("error generating thumbnail: %w", err)
		}
		thumb, err := saveThumbToFile(&thumbProps, source)
		if err != nil {
			return "", fmt.Errorf("error generating thumbnail: %w", err)
		}
		return thumb, nil
	case totalPixels <= 1920*1080:
		_, thumbProps, err := thumbnailer.Process(file, thumbnailer.Options{
			ThumbDims: thumbnailer.Dims{
				Width:  originalDims.Width / 6,
				Height: originalDims.Height / 6,
			},
		})
		if err != nil {
			return "", fmt.Errorf("error generating thumbnail: %w", err)
		}
		thumb, err := saveThumbToFile(&thumbProps, source)
		if err != nil {
			return "", fmt.Errorf("error generating thumbnail: %w", err)
		}
		return thumb, nil
	case totalPixels >= 320*240:
		_, thumbProps, err := thumbnailer.Process(file, thumbnailer.Options{
			ThumbDims: thumbnailer.Dims{
				Width:  originalDims.Width,
				Height: originalDims.Height,
			},
		})
		if err != nil {
			return "", fmt.Errorf("error generating thumbnail: %w", err)
		}
		thumb, err := saveThumbToFile(&thumbProps, source)
		if err != nil {
			return "", fmt.Errorf("error generating thumbnail: %w", err)
		}
		return thumb, nil
	default:
		aspectRatio := float64(originalDims.Width) / float64(originalDims.Height)
		var thumbWidth, thumbHeight uint
		if aspectRatio >= 1 { // Landscape or square videos
			thumbWidth = uint(aspectRatio * 480)
			thumbHeight = 480
		} else { // Portrait videos
			thumbWidth = 480
			thumbHeight = uint(480 / aspectRatio)
		}

		_, thumbProps, err := thumbnailer.Process(file, thumbnailer.Options{
			ThumbDims: thumbnailer.Dims{
				Width:  thumbWidth,
				Height: thumbHeight,
			},
		})
		if err != nil {
			return "", fmt.Errorf("error generating thumbnail: %w", err)
		}
		thumb, err := saveThumbToFile(&thumbProps, source)
		if err != nil {
			return "", fmt.Errorf("error generating thumbnail: %w", err)
		}
		return thumb, nil
	}
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

func (ss *StaticStorage) AddVideo(v *Video) error {
	thumb, err := generateThumbnail(v.Source)
	if err != nil {
		return fmt.Errorf("error adding video: %w", err)
	}
	ss.Log.Info().Msgf("Generated thumbnail for video %s, \nPath: %s", v.Source, thumb)

	_, err = ss.db.ExecContext(context.Background(), `INSERT INTO cdn.videos (source, thumbnail, alt)
		VALUES ($1, $2, $3)`, v.Source, thumb, v.Alt)
	if err != nil {
		return fmt.Errorf("error adding video: %w", err)
	}

	return nil
}
