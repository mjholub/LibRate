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
// @Success 200 {object} h.ResponseHTTP{}
// @Success 204 {object} h.ResponseHTTP{} "When the followee is already followed"
// @Failure 400 {object} h.ResponseHTTP{}
// @Failure 401 {object} h.ResponseHTTP{}
// @Failure 404 {object} h.ResponseHTTP{}
// @Failure 403 {object} h.ResponseHTTP{} "When at least one party blocks the other"
// @Failure 500 {object} h.ResponseHTTP{}
// @Router /members/follow [post]
func (mc *MemberController) Follow(c *fiber.Ctx) error {
	follower := c.Locals("jwtToken").(*jwt.Token).Claims.(jwt.MapClaims)["webfinger"].(string)
	var fr member.FollowBlockRequest
	if err := c.BodyParser(&fr); err != nil {
		return logAndRes(c, mc.log, err, 400, "Error parsing input")
	}
	fr.Requester = follower
	err := mc.storage.RequestFollow(c.Context(), &fr)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "sanitize"):
			return logAndRes(c, mc.log, err, 400, "Invalid input")
		case strings.Contains(err.Error(), sql.ErrNoRows.Error()):
			return logAndRes(c, mc.log, err, 404, "Member not found")
		case strings.Contains(err.Error(), "failed to"):
			return logAndRes(c, mc.log, err, 500, "Failed to request follow")
		case strings.Contains(err.Error(), "is blocked"):
			return logAndRes(c, mc.log, err, 403, "Blocked")
		case strings.Contains(err.Error(), "already followed"):
			return logAndRes(c, mc.log, err, 204, "Already following")
		default:
			return logAndRes(c, mc.log, err, 500, "Failed to request follow")
		}
	}

	if !isLocalRequest(follower, fr.Target) {
		return h.Res(c, fiber.StatusNotImplemented, "Remote follow not implemented")
	}

	return h.Res(c, 200, "Follow request sent")
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
// @Router /members/follow/requests/{id} [put]
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
// @Router /members/follow/requests/{id} [delete]
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
// @Param own query bool false "Get own sent follow requests. If false, get received follow requests"
// @Accept json
// @Success 200 {object} h.ResponseHTTP{data=[]int64}
// @Failure 400 {object} h.ResponseHTTP{}
// @Failure 401 {object} h.ResponseHTTP{}
// @Failure 404 {object} h.ResponseHTTP{}
// @Failure 500 {object} h.ResponseHTTP{}
// @Router /members/follow/requests [get]
func (mc *MemberController) GetFollowRequests(c *fiber.Ctx) error {
	webfinger := c.Locals("jwtToken").(*jwt.Token).Claims.(jwt.MapClaims)["webfinger"].(string)

	own := c.QueryBool("own", false)

	var requests []int64
	var err error

	if own {
		requests, err = mc.storage.GetFollowRequests(c.Context(), webfinger, true)
		if strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
			return logAndRes(c, mc.log, err, 404, "No follow requests found")
		}
		if err != nil {
			return logAndRes(c, mc.log, err, 500, "Failed to get follow requests")
		}
		return h.ResData(c, 200, "Follow requests", requests)
	} else {
		requests, err = mc.storage.GetFollowRequests(c.Context(), webfinger, false)
		if strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
			return logAndRes(c, mc.log, err, 404, "No follow requests found")
		}
		if err != nil {
			return logAndRes(c, mc.log, err, 500, "Failed to get follow requests")
		}
		return h.ResData(c, 200, "Sent follow requests", requests)
	}
}

// @Summary Unfollow a member
// @Description Unfollow a member or remove follower
// @Param Authorization header string true "The follower's JWT"
// @Param followee body string true "The webfinger of the member to unfollow"
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

func logAndRes(c *fiber.Ctx, logger *zerolog.Logger, err error, status int, msg string) error {
	logger.Error().Err(err).Msg(msg)
	return h.Res(c, status, msg)
}

func isLocalRequest(sender, recipient string) bool {
	senderDomain := strings.Split(sender, "@")[1]
	recipientDomain := strings.Split(recipient, "@")[1]
	return senderDomain == recipientDomain
}
