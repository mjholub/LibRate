package static

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/golang-jwt/jwt/v5"

	h "codeberg.org/mjh/LibRate/internal/handlers"
)

// UploadImage takes the nickname of the uploader (always), image type:
// "profile", "banner", "album_cover"
// image type is used to construct the image path.
// The image is saved to the filesystem, and the path is saved to the database.
// This handler returns the ID of the image in the database.
// When a profile picture/banner upload is requested, it still needs to be confirmed by the user.
// WARN: this is not concurrency-safe yet
func (s *StaticController) UploadImage(c *fiber.Ctx) error {
	member := c.Locals("jwtToken").(*jwt.Token)
	claims := member.Claims.(jwt.MapClaims)
	name := claims["member_name"].(string)
	memberName := c.FormValue("member")
	if name != memberName {
		return fiber.ErrForbidden
	}
	imageType := c.FormValue("imageType")
	file, err := c.FormFile("fileData")
	if err != nil {
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

	savePath, imageID, err := s.storage.AddImage(c.UserContext(), imageType, ext, memberName, nil)
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

	return h.ResData(c, 201, "Success", fiber.Map{
		"pic_id": imageID,
	})
}

func (s *StaticController) DeleteImage(c *fiber.Ctx) error {
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
