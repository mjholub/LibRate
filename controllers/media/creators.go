package media

import (
	"fmt"
	"strconv"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"

	h "codeberg.org/mjh/LibRate/internal/handlers"
)

func (mc *Controller) GetCreatorByID(c *fiber.Ctx) error {
	if c.Query("kind") == "" || c.Query("id") == "" {
		return h.Res(c, fiber.StatusBadRequest, "Missing kind or ID")
	}
	id := c.Query("id")
	if c.Query("kind") == "person" {
		idInt, err := i64fromID(id)
		if err != nil {
			return h.Res(c, fiber.StatusBadRequest, "Invalid ID: "+id)
		}
		creator, err := mc.storage.Ps.GetPerson(c.UserContext(), idInt)
		if err != nil {
			return h.Res(c, fiber.StatusInternalServerError, fmt.Sprintf("Failed to get creator with ID %s: %s", id, err.Error()))
		}

		creatorJSON, err := json.Marshal(creator)
		if err != nil {
			return h.Res(c, fiber.StatusInternalServerError, "Failed to marshal creator: "+err.Error())
		}

		return h.Res(c, fiber.StatusOK, string(creatorJSON))
	} else if c.Query("kind") == "group" {
		idInt, err := i64fromID(id)
		if err != nil {
			return h.Res(c, fiber.StatusBadRequest, "Invalid ID: "+id)
		}
		creator, err := mc.storage.Ps.GetGroup(c.UserContext(), int32(idInt))
		if err != nil {
			return h.Res(c, fiber.StatusInternalServerError, "Failed to get creator: "+err.Error())
		}

		creatorJSON, err := json.Marshal(creator)
		if err != nil {
			return h.Res(c, fiber.StatusInternalServerError, "Failed to marshal creator: "+err.Error())
		}

		return h.Res(c, fiber.StatusOK, string(creatorJSON))
	}
	return h.Res(c, fiber.StatusBadRequest, "Invalid kind: "+c.Query("kind"))
}

// @Summary Get the cast of the media with given ID
// @Description Get the full cast and crew involved with the creation of the media with given ID
// @Tags media,artists,bulk operations,films,television,anime
// @Accept json
// @Produce json
// @Param media_id path string true "The UUID of the media to get the cast of"
// @Success 200 {object} h.ResponseHTTP{data=media.Cast}
// @Failure 400 {object} h.ResponseHTTP{}
// @Failure 500 {object} h.ResponseHTTP{}
// @Router /media/{media_id}/cast [get]
func (mc *Controller) GetCastByMediaID(c *fiber.Ctx) error {
	// get the media ID and then query the junction tables in the database and perform a join into a JSON corresponding to the Cast type
	if c.Params("media_id") == "" {
		return h.Res(c, fiber.StatusBadRequest, "Missing ID")
	}
	id := c.Params("media_id")
	mediaID, err := uuid.FromString(id)
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, fmt.Sprintf("Invalid ID: %s (%v)", id, err.Error()))
	}
	cast, err := mc.storage.GetCast(c.UserContext(), mediaID)
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, "Failed to get cast: "+err.Error())
	}
	castJSON, err := json.Marshal(cast)
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, "Failed to marshal cast: "+err.Error())
	}
	return h.Res(c, fiber.StatusOK, string(castJSON))
}

func i64fromID(id string) (int64, error) {
	idInt, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to parse ID %s", id)
	}
	return idInt, nil
}
