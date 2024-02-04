package media

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/rs/zerolog"
	"github.com/samber/lo"

	"codeberg.org/mjh/LibRate/cfg"
	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/models/media"
)

type (
	// IController is the interface for the media controller
	// It defines the methods that the media controller must implement
	// This is useful for mocking the media controller in unit tests
	IController interface {
		GetMedia(c *fiber.Ctx) error
		GetRandom(c *fiber.Ctx) error
		AddMedia(c *fiber.Ctx) error
	}

	// Controller is the controller for media endpoints
	// The methods which are the receivers of this struct are a bridge between the fiber layer and the storage layer
	Controller struct {
		storage media.Storage
		conf    *cfg.Config
	}

	mediaError struct {
		ID  uuid.UUID
		Err error
	}
)

func NewController(storage media.Storage, conf *cfg.Config) *Controller {
	return &Controller{storage: storage, conf: conf}
}

// GetMedia retrieves media information based on the media ID
// `media ID` is a UUID (binary, but passed from the frontend as a string)
// This might be changed in the future, most likely to uint8array
// TODO: modify the response to utilize generics instead of interface{}
// @Summary Retrieve media information
// @Description Retrieve complete media information for the given media ID
// @Tags media,metadata
// @Param id path string true "Media UUID"
// @Accept json
// @Produce json
// @Success 200 {object} h.ResponseHTTP{data=any}
// @Failure 400 {object} h.ResponseHTTP{}
// @Failure 500 {object} h.ResponseHTTP{}
// @Router /media/{id} [get]
func (mc *Controller) GetMedia(c *fiber.Ctx) error {
	mediaID, err := uuid.FromString(c.Params("id"))
	if err != nil {
		mc.storage.Log.Error().Err(err).
			Msgf("Failed to parse media ID %s", c.Params("id"))
		return h.Res(c, fiber.StatusBadRequest, "Invalid media ID")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	media, err := mc.storage.Get(ctx, mediaID)
	if err != nil {
		mc.storage.Log.Error().Err(err).
			Msgf("Failed to get media with ID %s", c.Params("id"))
		return h.Res(c, fiber.StatusInternalServerError, "Failed to get media")
	}

	detailedMedia, err := mc.storage.
		GetDetails(ctx, media.Kind, media.ID)
	if err != nil {
		mc.storage.Log.Error().Err(err).Msgf("Failed to get media details for media with ID %s: %v", c.Params("id"), err)
		return h.Res(c, fiber.StatusInternalServerError, "Failed to get media details")
	}

	return h.ResData(c, fiber.StatusOK, "success", detailedMedia)
}

// GetImagePaths retrieves the image paths for the media with the given ID
// @Summary Retrieve image paths
// @Description Retrieve the image paths for the media with the given ID
// @Tags media,metadata,images
// @Param media_id path string true "Media UUID"
// @Accept json text/plain
// @Produce text/plain
// @Param media_id path string true "Media UUID"
// @Success 200 {string} string "Image path"
// @Failure 400 {object} h.ResponseHTTP{}
// @Failure 404 {object} h.ResponseHTTP{}
// @Failure 500 {object} h.ResponseHTTP{}
// @Router /media/{media_id}/images [get]
func (mc *Controller) GetImagePaths(c *fiber.Ctx) error {
	mc.storage.Log.Info().Msg("Hit endpoint " + c.Path())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mediaID, err := uuid.FromString(c.Params("media_id"))
	if err != nil {
		return handleBadRequest(mc.storage.Log, c, "Invalid media ID")
	}

	kind, err := mc.storage.GetKind(ctx, mediaID)
	if err != nil {
		return handleInternalError(mc.storage.Log, c, "Failed to get kind", err)
	}
	mc.storage.Log.Debug().Msgf("Got kind %s for media with ID %s", kind, c.Params("media_id"))

	path, err := mc.storage.GetImagePath(ctx, mediaID)
	if err == sql.ErrNoRows {
		return handlePlaceholderImage(mc.storage.Log, c, kind)
	} else if err != nil {
		return handleInternalError(mc.storage.Log, c, "Failed to get image paths", err)
	}

	return handleImageResponse(mc.storage.Log, c, path)
}

// GetGenres retrieves the genres for the given genre kind
// @Summary Retrieve genres
// @Description Retrieve the list of genres of the specified type
// @Tags media,genres,bulk operations
// @Param kind path string true "Genre kind" Enums(film, tv, music, book, game)
// @Param names_only query bool false "Return only genre names. Usually used for populating dropdowns"
// @Param as_links query bool false "Return the genre names as links"
// @Param all query bool false "Return all genres, not only the ones without a parent genre (e.g. Twee Pop and Jangle Pop instead of just Pop)"
// @Param columns query []string false "Return only the specified columns" Enums(name, id, kinds, parent, children)
// @default names
// @example kind=film&names_only=true&as_links=true&all=true&columns=name,id,kinds,parent,children
// @Accept json
// @Produce json
// @Success 200 {object} h.ResponseHTTP{data=[]string} "If names_only or as_links=true"
// @Success 200 {object} h.ResponseHTTP{data=[]models.Genre} "If names_only=false and as_links=false"
// @Failure 400 {object} h.ResponseHTTP{}
// @Failure 500 {object} h.ResponseHTTP{}
// @Router /genres/{kind} [get]
func (mc *Controller) GetGenres(c *fiber.Ctx) error {
	genreKind := c.Params("kind")
	namesOnly := c.QueryBool("names_only", true)
	possible := []string{"film", "tv", "music", "book", "game"}
	if !lo.Contains(possible, genreKind) {
		return handleBadRequest(mc.storage.Log, c, "Invalid genre kind")
	}

	asLinks := c.QueryBool("as_links", false)
	if !namesOnly && asLinks {
		mc.storage.Log.Warn().Msg("correcting incorrect query param combination names_only=false and as_links=true")
		namesOnly = true
	}
	all := c.QueryBool("all", true)

	if namesOnly {
		mc.storage.Log.Debug().Msgf("Getting genre names for %s", genreKind)
		genreNames, err := media.GetGenres[[]string](&mc.storage, c.Context(), genreKind, all, "name")
		if err != nil {
			return handleInternalError(mc.storage.Log, c, "Failed to get genre names", err)
		}
		if asLinks {
			genreLinks := h.LinksFromArray(fmt.Sprintf("%s/genres/%s", c.Path(), genreKind), genreNames)
			return c.JSON(genreLinks)
		}
		return c.JSON(genreNames)
	}

	columns := c.Query("columns", "name")

	mc.storage.Log.Debug().Msgf("Getting following columns for %s: %s", genreKind, columns)
	genres, err := media.GetGenres[[]media.Genre](&mc.storage, c.Context(), genreKind, all, columns)
	if err != nil {
		return handleInternalError(mc.storage.Log, c, "Failed to get genres", err)
	}
	return h.ResData(c, fiber.StatusOK, "success", genres)
}

// GetGenre retrieves a single genre
// @Summary Retrieve genre
// @Description Retrieve the genre with the given name and type
// @Tags media,genres
// @Param kind path string true "Genre kind" Enums(film, tv, music, book, game)
// @Param genre path string true "Genre name (snake_lowercase)"
// @Param lang query string false "ISO-639-1 language code" Enums(en, de)
// @Accept json
// @Produce json
// @Success 200 {object} h.ResponseHTTP{data=models.Genre}
// @Failure 400 {object} h.ResponseHTTP{}
// @Failure 404 {object} h.ResponseHTTP{}
// @Failure 500 {object} h.ResponseHTTP{}
// @Router /genre/{kind}/{genre} [get] "Note the singular genre, not genres"
func (mc *Controller) GetGenre(c *fiber.Ctx) error {
	genreKind := c.Params("kind")
	possible := []string{"film", "tv", "music", "book", "game"}
	if !lo.Contains(possible, genreKind) {
		return handleBadRequest(mc.storage.Log, c, "Invalid genre kind")
	}
	lang := c.Query("lang", "en")

	genreName := c.Params("genre")
	genre, err := mc.storage.GetGenre(c.Context(), genreKind, lang, genreName)
	if err != nil && strings.Contains(err.Error(), "no rows in result set") {
		return h.Res(c, fiber.StatusNotFound, "Genre not found")
	}
	if err != nil {
		return handleInternalError(mc.storage.Log, c, "Failed to get genre", err)
	}
	return h.ResData(c, fiber.StatusOK, "success", genre)
}

// GetArtistsByName is a POST endpoint that takes the list of artists as a multipart form data
// and returns the artists with their IDs as a response
// @Summary Retrieve artists
// @Description Retrieve the artists with the given names
// @Tags media,artists,bulk operations
// @Param names formData []string true "Artist names"
// @Accept multipart/form-data
// @Produce json
// @Success 200 {object} h.ResponseHTTP{data=models.GroupedArtists}
// @Failure 400 {object} h.ResponseHTTP{}
// @Failure 500 {object} h.ResponseHTTP{}
// @Router /artists/by-name [post]
func (mc *Controller) GetArtistsByName(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}

func handleBadRequest(log *zerolog.Logger, c *fiber.Ctx, message string) error {
	log.Error().Msgf("Failed to %s", message)
	return h.Res(c, fiber.StatusBadRequest, message)
}

func handleInternalError(log *zerolog.Logger, c *fiber.Ctx, message string, err error) error {
	log.Error().Err(err).Msg(message)
	return h.Res(c, fiber.StatusInternalServerError, message)
}

func handlePlaceholderImage(log *zerolog.Logger, c *fiber.Ctx, kind string) error {
	log.Warn().Msgf("Using placeholder image for media with ID %s", c.Params("id"))
	var placeholderPath string
	switch kind {
	case "film", "tv_show":
		placeholderPath = "./static/video/placeholder.svg"
	case "album", "track":
		placeholderPath = "./static/music/placeholder.webp"
	default:
		placeholderPath = "./static/placeholder.png"
	}

	err := c.SendString(placeholderPath)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to send placeholder image for media with ID %s", c.Params("id"))
		return h.Res(c, fiber.StatusNotFound, "Failed to send placeholder image")
	}
	return c.SendStatus(fiber.StatusOK)
}

func handleImageResponse(log *zerolog.Logger, c *fiber.Ctx, path string) error {
	log.Debug().Msgf("Got image path %s for media with ID %s", path, c.Params("id"))
	err := c.SendString("./static/" + path)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to send image for media with ID %s", c.Params("id"))
		return h.Res(c, fiber.StatusNotFound, "Failed to send image")
	}
	return c.SendStatus(fiber.StatusOK)
}
