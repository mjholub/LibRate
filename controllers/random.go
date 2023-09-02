package controllers

import (
	"context"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"

	h "codeberg.org/mjh/LibRate/internal/handlers"
)

// GetRandom fetches up to 5 random media items to be displayed in a carousel on the home page
func (mc *MediaController) GetRandom(c *fiber.Ctx) error {
	mc.storage.Log.Info().Msg("Hit endpoint " + c.Path())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// FIXME: blacklist tracks, because they get e.g. improperly parsed as albums at some point
	media, err := mc.storage.GetRandom(ctx, 2, "track")
	if err != nil {
		mc.storage.Log.Error().Err(err).Msgf("Failed to get random media: %s", err.Error())
		return h.Res(c, fiber.StatusInternalServerError,
			"Failed to get random media: "+err.Error())
	}

	mc.storage.Log.Info().Msgf("Got %d random media items", len(media))
	mediaItems := make([]interface{}, len(media))

	errChan := make(chan mediaError, len(media))
	var wg sync.WaitGroup

	// set an iterator variable so we can access the media item in the goroutine
	i := 0
	mc.storage.Log.Info().Msg("Getting media details")
	for id, kind := range media {
		wg.Add(1)
		go func(i int, id uuid.UUID, kind string) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			// NOTE:this val. of err must be shadowed, otherwise it will be the same err for all goroutines
			mDetails, err := mc.storage.
				GetMediaDetails(ctx, kind, id)
			if err != nil {
				errChan <- mediaError{ID: id, Err: err}
				return
			}
			mc.storage.Log.Info().Msgf("Got media details for media with ID %s", id.String())
			mediaItems[i] = mDetails
		}(i, id, kind)
		i++
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		for e := range errChan {
			mc.storage.Log.Error().Err(err).
				Msgf("Failed to get media details for media with ID %s: %s", e.ID, e.Err.Error())
		}
		return h.Res(c, fiber.StatusInternalServerError, "Failed to get media details")
	}

	return h.ResData(c, fiber.StatusOK, "success", mediaItems)
}