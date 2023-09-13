package members

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	h "codeberg.org/mjh/LibRate/internal/handlers"
)

// GetMember retrieves user information based on the user ID
func (mc *MemberController) GetMember(c *fiber.Ctx) error {
	mc.log.Info().Msg("GetMember called")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mc.log.Debug().Msgf("ID: %s", c.Params("id"))
	member, err := mc.storage.Read(ctx, "id", c.Params("id"))
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, "Member not found")
	}
	mc.log.Info().Msgf("Member: %+v", member)

	return h.ResData(c, fiber.StatusOK, "success", member)
}

func (mc *MemberController) GetMemberByNick(c *fiber.Ctx) error {
	mc.log.Info().Msg("GetMemberByNick called")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mc.log.Debug().Msgf("Nick: %s", c.Params("nickname"))
	member, err := mc.storage.Read(ctx, "nick", c.Params("nickname"))
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, "Member not found")
	}
	mc.log.Info().Msgf("Member: %+v", member)

	return h.ResData(c, fiber.StatusOK, "success", member)
}
