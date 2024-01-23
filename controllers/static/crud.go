package static

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"

	"github.com/chai2010/webp"
	"github.com/golang-jwt/jwt/v5"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/internal/lib/thumbnailer"
	"codeberg.org/mjh/LibRate/models/static"
)

// UploadImage takes the nickname of the uploader (always), image type:
// "profile", "banner", "album_cover"
// image type is used to construct the image path.
// The image is saved to the filesystem, and the path is saved to the database.
// This handler returns the ID of the image in the database.
// When a profile picture/banner upload is requested, it still needs to be confirmed by the user.
// INFO: currently supported arguments for imageType: "profile", "album_cover"
func (s *Controller) UploadImage(c *fiber.Ctx) error {
	member := c.Locals("jwtToken").(*jwt.Token)
	claims := member.Claims.(jwt.MapClaims)
	name := claims["member_name"].(string)
	memberName := c.FormValue("member")
	if name != memberName {
		s.log.Warn().Msgf("Member %s tried to upload an image for %s", name, memberName)
		return fiber.ErrForbidden
	}
	imageType := c.FormValue("imageType")
	file, err := c.FormFile("fileData")
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to get file from form")
		return fiber.ErrBadRequest
	}

	if file.Size > s.conf.Fiber.MaxUploadSize {
		s.log.Warn().Msgf("File too large: %d", file.Size)
		return fiber.ErrRequestEntityTooLarge
	}
	// if MIME != "image/*"
	if file.Header.Get("Content-Type")[:5] != "image" {
		s.log.Warn().Msgf("Invalid MIME type: %s", file.Header.Get("Content-Type"))
		return fiber.ErrUnsupportedMediaType
	}

	split := strings.Split(file.Filename, ".")
	ext := split[len(split)-1]

	var savePath string
	var imageID int64
	if imageType == "profile" {
		savePath, imageID, err = s.saveProfileImage(c.UserContext(), memberName, ext, file)
	} else {
		props := static.MediaProps{
			Uploader:  memberName,
			Ext:       ext,
			ImageType: imageType,
		}
		savePath, imageID, err = s.storage.AddImage(c.UserContext(), &props)
	}
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to add image")
		return fiber.ErrInternalServerError
	}

	s.log.Debug().Msgf(`Image ID: %d. Original name: %s. MIME type: %s. Size: %d. Saving to: %s...`,
		imageID, file.Filename, file.Header.Get("Content-Type"), file.Size, savePath)

	if err := c.SaveFile(file, savePath); err != nil {
		s.log.Error().Err(err).Msg("Failed to save image")
		return fiber.ErrInternalServerError
	}

	go s.saveThumb(imageType, savePath, file.Header.Get("Content-Type"))
	// flush the cache
	// e.g. for static/img/profile/lain.png
	// redis caches "static/img/profile/lain.png_GET_body"
	if err := s.cache.Delete(savePath + "_GET_body"); err != nil {
		s.log.Warn().Err(err).Msgf("Failed to flush cache for %s", savePath)
	}

	return h.ResData(c, 201, "Success", fiber.Map{
		"pic_id": imageID,
	})
}

// we are using the recover middleware, so it's relatively safe to call this in a goroutine
// without returning an error, since goroutine callback would necessitate a blocking call to wait on reading from
// error channel
func (s *Controller) saveThumb(
	imageType,
	imagePath,
	imageFormat string,
) {
	var dims *thumbnailer.Dims
	ns := s.conf.Fiber.Thumbnailing.TargetNS
	for i := range ns {
		// #wontfix providing more than one ns and all simultaneously
		if lo.Contains(ns[i].Names, imageType) || ns[i].Names[0] == "all" {
			dims = &ns[i].MaxSize
			break
		}
	}

	if dims == nil {
		s.log.Warn().Msgf(
			`Image of type (as in use case) %s was provided,
		but no thumbnailing namespace was found for it.`,
			imageType,
		)
		return
	}

	imagePathBase := strings.TrimSuffix(imagePath, filepath.Ext(imagePath))
	imgPathNew := imagePathBase + "_thumb" + filepath.Ext(imagePath)
	f, err := os.Create(imgPathNew)
	if err != nil {
		s.log.Error().Err(err).Msgf("Failed to save thumbnail for image %s", imagePath)
		return
	}
	defer f.Close()

	thumb, err := thumbnailer.Thumbnail(*dims, imagePath)
	if err != nil {
		s.log.Error().Err(err).Msgf("Failed to create thumbnail for image %s", imagePath)
		return
	}
	switch imageFormat {
	case "image/jpeg":
		if err := jpeg.Encode(f, thumb, nil); err != nil {
			s.log.Error().Err(err).Msgf("Failed to encode thumbnail for image %s", imagePath)
			return
		}
	case "image/png":
		if err := png.Encode(f, thumb); err != nil {
			s.log.Error().Err(err).Msgf("Failed to encode thumbnail for image %s", imagePath)
			return
		}
	case "image/gif":
		if err := gif.Encode(f, thumb, nil); err != nil {
			s.log.Error().Err(err).Msgf("Failed to encode thumbnail for image %s", imagePath)
			return
		}
	case "image/webp":
		opts := &webp.Options{
			Lossless: false,
			Quality:  80,
		}
		if err := webp.Encode(f, thumb, opts); err != nil {
			s.log.Error().Err(err).Msgf("Failed to encode thumbnail for image %s", imagePath)
			return
		}
	default:
		s.log.Error().Err(err).Msgf("Failed to encode thumbnail for image %s for MIME %s", imagePath, imageType)
		return
	}
}

func (s *Controller) saveProfileImage(
	ctx context.Context,
	memberName, ext string,
	file *multipart.FileHeader,
) (string, int64, error) {
	// check if the image already exists
	hash, err := calculateHash(file)
	if err != nil {
		s.log.Error().Err(err).Msgf("Failed to calculate hash for image with name:%s, uploaded by %s: %v", file.Filename, memberName, err)
		return "", 0, fiber.ErrInternalServerError
	}
	id, err := s.storage.LookupHash(ctx, hash, memberName)
	if err != nil && err != sql.ErrNoRows {
		s.log.Error().Err(err).
			Msgf("Failed to lookup hash for image with name:%s, uploaded by %s",
				file.Filename, memberName)
		return "", 0, fiber.ErrInternalServerError
	}

	if id != 0 {
		s.log.Debug().Msgf("Image with hash %s already exists", hash)
		return "", 0, fiber.ErrConflict
	}

	props := static.MediaProps{
		Uploader:  memberName,
		Ext:       ext,
		Hash:      hash,
		ImageType: "profile",
	}

	return s.storage.AddImage(ctx, &props)
}

func (s *Controller) DeleteImage(c *fiber.Ctx) error {
	// 1. get the image id
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return h.Res(c, 400, "Invalid image ID")
	}
	// 2. get the jwt claims
	member := c.Locals("jwtToken").(*jwt.Token)
	claims := member.Claims.(jwt.MapClaims)
	name := claims["member_name"].(string)
	// 3. get the image owner from the database
	owner, err := s.storage.GetOwner(c.UserContext(), int64(id))
	if err == context.Canceled {
		return fiber.ErrRequestTimeout
	}
	if name != owner {
		return fiber.ErrForbidden
	}

	// 4. delete the image from the filesystem and database
	var path string
	path, err = s.storage.DeleteImage(c.UserContext(), int64(id))
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to delete image with id " + idStr)
		return fiber.ErrInternalServerError
	}

	if err := os.Remove(path); err != nil {
		s.log.Error().Err(err).Msgf("Failed to delete image stored at %s: %v", path, err)
		return fiber.ErrInternalServerError
	}

	return h.Res(c, 200, "Success")
}

func calculateHash(file *multipart.FileHeader) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}

	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	hashSum := hasher.Sum(nil)

	return hex.EncodeToString(hashSum), nil
}
