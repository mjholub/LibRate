package members

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/models/member"
)

// @Summary Send a follow request to a member (user)
// @Description Send a follow request to another user. If the user has automatic follow request acceptance
// @Description enabled, the follow request will be accepted immediately.
// @Param Authorization header string true "The requester's JWT. Contains encrypted claims to the webfinger"
// @Param followee body string true "The webfinger of the member to follow"
// @Param notify body bool false "Receive notifications for all content created by the followee"
// @Param reblogs body bool true "Show this account's reblogs in home timeline"
// @Accept json
// @Success 200 {object} h.ResponseHTTP{data=member.FollowResponse}
// @Success 204 {object} h.ResponseHTTP{} "When the followee is already followed"
// @Failure 400 {object} h.ResponseHTTP{}
// @Failure 401 {object} h.ResponseHTTP{}
// @Failure 404 {object} h.ResponseHTTP{}
// @Failure 403 {object} h.ResponseHTTP{} "When at least one party blocks the other"
// @Failure 500 {object} h.ResponseHTTP{}
// @Router /members/follow [post]
func (mc *MemberController) Follow(c *fiber.Ctx) error {
	follower := c.Locals("jwtToken").(*jwt.Token).Claims.(jwt.MapClaims)["webfinger"].(string)
	mc.log.Debug().Msgf("Follower: %s", follower)
	var fr member.FollowBlockRequest
	if err := c.BodyParser(&fr); err != nil {
		return logAndRes(c, mc.log, err, 400, "Error parsing input")
	}
	fr.Requester = follower
	mc.log.Debug().Msgf("parsed the request into: %+v", fr)

	resp := mc.storage.RequestFollow(c.Context(), &fr)
	switch resp.Status {
	case "not_found":
		return logAndRes(c, mc.log, resp.Error, 404, "Member not found")
	case "failed":
		return logAndRes(c, mc.log, resp.Error, 500, "Failed to request follow")
	case "blocked":
		return logAndRes(c, mc.log, resp.Error, 403, "Blocked")
	case "already_following":
		return logAndRes(c, mc.log, resp.Error, 204, "Already following")
	default:
		if !isLocalRequest(follower, fr.Target) {
			return h.Res(c, fiber.StatusNotImplemented, "Remote follow not implemented")
		}

		return h.ResData(c, 200, "Follow request sent", resp)
	}
}

// @Summary Accept a follow request
// @Description Accept a follow request
// @Param Authorization header string true "The accepter's JWT"
// @Param id body int64 true "The ID of the follow request"
// @Accept json
// @Success 200 {object} h.ResponseHTTP{}
// @Failure 400 {object} h.ResponseHTTP{}
// @Failure 401 {object} h.ResponseHTTP{}
// @Failure 404 {object} h.ResponseHTTP{}
// @Failure 500 {object} h.ResponseHTTP{}
// @Router /members/follow/requests/in/{id} [put]
func (mc *MemberController) AcceptFollow(c *fiber.Ctx) error {
	accepter := c.Locals("jwtToken").(*jwt.Token).Claims.(jwt.MapClaims)["webfinger"].(string)
	var id int64
	var err error
	ID := c.Params("id")
	if id, err = strconv.ParseInt(ID, 10, 64); err != nil {
		return logAndRes(c, mc.log, err, 400, "Invalid input")
	}
	err = mc.storage.AcceptFollow(c.Context(), accepter, id)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "sanitize"):
			return logAndRes(c, mc.log, err, 400, "Invalid input")
		case strings.Contains(err.Error(), sql.ErrNoRows.Error()):
			return logAndRes(c, mc.log, err, 404, "Member not found")
		case strings.Contains(err.Error(), "failed to"):
			return logAndRes(c, mc.log, err, 500, "Failed to accept follow")
		case strings.Contains(err.Error(), "is blocked"):
			return logAndRes(c, mc.log, err, 403, "Blocked")
		case strings.Contains(err.Error(), "does not belong to"):
			return logAndRes(c, mc.log, err, 403, "Not your follow request")
		case strings.Contains(err.Error(), "already followed"):
			return logAndRes(c, mc.log, err, 204, "Already following")
		default:
			return logAndRes(c, mc.log, err, 500, "Failed to accept follow")
		}
	}

	return h.Res(c, 200, "Follow request accepted")
}

// @Summary Reject a follow request
// @Description Reject a follow request
// @Param Authorization header string true "The rejecter's JWT"
// @Param follower body string true "The webfinger of the member who requested to follow"
// @Param id path int64 true "The ID of the follow request"
// @Accept json
// @Success 200 {object} h.ResponseHTTP{}
// @Failure 400 {object} h.ResponseHTTP{}
// @Failure 401 {object} h.ResponseHTTP{}
// @Failure 404 {object} h.ResponseHTTP{}
// @Failure 500 {object} h.ResponseHTTP{}
// @Router /members/follow/requests/in/{id} [delete]
func (mc *MemberController) RejectFollow(c *fiber.Ctx) error {
	rejecter := c.Locals("jwtToken").(*jwt.Token).Claims.(jwt.MapClaims)["webfinger"].(string)
	var id int64
	var err error
	ID := c.Params("id")
	if id, err = strconv.ParseInt(ID, 10, 64); err != nil {
		return logAndRes(c, mc.log, err, 400, "Invalid input")
	}
	err = mc.storage.RejectFollow(c.Context(), rejecter, id)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "sanitize"):
			return logAndRes(c, mc.log, err, 400, "Invalid input")
		case strings.Contains(err.Error(), sql.ErrNoRows.Error()):
			return logAndRes(c, mc.log, err, 404, "Member not found")
		case strings.Contains(err.Error(), "failed to"):
			return logAndRes(c, mc.log, err, 500, "Failed to reject follow")
		case strings.Contains(err.Error(), "is blocked"):
			return logAndRes(c, mc.log, err, 403, "Blocked")
		case strings.Contains(err.Error(), "does not belong to"):
			return logAndRes(c, mc.log, err, 403, "Not your follow request")
		case strings.Contains(err.Error(), "already followed"):
			return logAndRes(c, mc.log, err, 204, "Already following")
		default:
			return logAndRes(c, mc.log, err, 500, "Failed to reject follow")
		}
	}

	return h.Res(c, 200, "Follow request rejected")
}

// @Summary Get follow requests
// @Description Get own received follow requests or sent follow requests
// @Param Authorization header string true "The JWT of the member. Contains encrypted claims to webfinger"
// @Param type path string true "Which follow requests should be looked up" Enums(sent, received, all)
// @Accept json
// @Success 200 {object} h.ResponseHTTP{data=[]int64}
// @Failure 400 {object} h.ResponseHTTP{}
// @Failure 401 {object} h.ResponseHTTP{}
// @Failure 404 {object} h.ResponseHTTP{}
// @Failure 500 {object} h.ResponseHTTP{}
// @Router /members/follow/requests [get]
func (mc *MemberController) GetFollowRequests(c *fiber.Ctx) error {
	webfinger := c.Locals("jwtToken").(*jwt.Token).Claims.(jwt.MapClaims)["webfinger"].(string)

	requestsType := c.Params("type")

	requests, err := mc.storage.GetFollowRequests(c.Context(), webfinger, requestsType)
	if err != nil {
		return h.BadRequest(mc.log, c, "error fetching follow requests for type"+requestsType,
			"failed to load follow requests", err)
	}

	return h.ResData(c, 200, "follow requests:", requests)
}

// @Summary Unfollow a member
// @Description Unfollow a member or remove follower
// @Param Authorization header string true "The follower's JWT"
// @Param target body string true "The webfinger of the member to unfollow"
// @Accept json
// @Success 200 {object} h.ResponseHTTP{}
// @Failure 400 {object} h.ResponseHTTP{}
// @Failure 401 {object} h.ResponseHTTP{}
// @Failure 404 {object} h.ResponseHTTP{}
// @Failure 500 {object} h.ResponseHTTP{}
// @Router /members/follow [delete]
func (mc *MemberController) Unfollow(c *fiber.Ctx) error {
	follower := c.Locals("jwtToken").(*jwt.Token).Claims.(jwt.MapClaims)["webfinger"].(string)
	var fr member.FollowBlockRequest
	if err := c.BodyParser(&fr); err != nil {
		return logAndRes(c, mc.log, err, 400, "Error parsing input")
	}
	fr.Requester = follower
	err := mc.storage.RemoveFollower(c.Context(), follower, fr.Target)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "sanitize"):
			return logAndRes(c, mc.log, err, 400, "Invalid input")
		case strings.Contains(err.Error(), sql.ErrNoRows.Error()):
			return logAndRes(c, mc.log, err, 404, "Member not found")
		case strings.Contains(err.Error(), "failed to"):
			return logAndRes(c, mc.log, err, 500, "Failed to unfollow")
		case strings.Contains(err.Error(), "is blocked"):
			return logAndRes(c, mc.log, err, 403, "Blocked")
		case strings.Contains(err.Error(), "already followed"):
			return logAndRes(c, mc.log, err, 204, "Already following")
		default:
			return logAndRes(c, mc.log, err, 500, "Failed to unfollow")
		}
	}

	if !isLocalRequest(follower, fr.Target) {
		return h.Res(c, fiber.StatusNotImplemented, "Remote follow not implemented")
	}

	return h.Res(c, 200, "Unfollowed")
}

// @Summary Check if a member is followed
// @Description Check if a member is followed by the request initiator
// @Param Authorization header string true "The follower's JWT"
// @Param followee_webfinger path string true "The webfinger of the member to check"
// @Accept json
// @Produce json
// @Success 200 {object} h.ResponseHTTP{data=member.FollowResponse}
// @Failure 400 {object} h.ResponseHTTP{}
// @Failure 401 {object} h.ResponseHTTP{}
// @Failure 404 {object} h.ResponseHTTP{}
// @Failure 500 {object} h.ResponseHTTP{}
// @Router /members/follow/status/{followee_webfinger} [get]
func (mc *MemberController) FollowStatus(c *fiber.Ctx) error {
	follower := c.Locals("jwtToken").(*jwt.Token).Claims.(jwt.MapClaims)["webfinger"].(string)
	followee := c.Params("followee_webfinger")
	resp := mc.storage.GetFollowStatus(c.Context(), follower, followee)
	if resp.Status == "failed" {
		mc.log.Error().Err(resp.Error).Msg("Failed to get follow status")
		return h.Res(c, fiber.StatusInternalServerError, "Failed to get follow status")
	}

	return h.ResData(c, 200, "Follow status", resp)
}

func logAndRes(c *fiber.Ctx, logger *zerolog.Logger, err error, status int, msg string) error {
	logger.Error().Err(err).Msg(msg)
	return h.Res(c, status, msg)
}

func isLocalRequest(sender, recipient string) bool {
	senderDomain := strings.Split(sender, "@")[1]
	recipientDomain := strings.Split(recipient, "@")[1]
	return senderDomain == recipientDomain
}

// @Summary Cancel a follow request
// @Description Cancel a sent follow request if it's pending
// @Param Authorization header string true "The requester's JWT"
// @Param X-CSRF-Token header string true "The CSRF token"
// @Param id path int64 true "The ID of the follow request"
// @Accept json
// @Success 200 {object} h.ResponseHTTP{}
// @Failure 400 {object} h.ResponseHTTP{}
// @Failure 401 {object} h.ResponseHTTP{}
// @Failure 404 {object} h.ResponseHTTP{}
// @Failure 500 {object} h.ResponseHTTP{}
// @Router /members/follow/requests/out/{id} [delete]
func (mc *MemberController) CancelFollowRequest(c *fiber.Ctx) error {
	requester := c.Locals("jwtToken").(*jwt.Token).Claims.(jwt.MapClaims)["webfinger"].(string)
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return h.BadRequest(mc.log, c, "invalid input", "invalid input", err)
	}
	err = mc.storage.CancelFollow(c.Context(), requester, id)
	switch {
	case strings.Contains(err.Error(), "failed to "):
		return h.BadRequest(mc.log, c, "failed to cancel follow request", "failed to cancel follow request", err)
	case strings.Contains(err.Error(), "does not belong to "):
		return h.Res(c, fiber.StatusForbidden, "Not your follow request")
	default:
		return h.Res(c, fiber.StatusOK, "Follow request cancelled")
	}
}
