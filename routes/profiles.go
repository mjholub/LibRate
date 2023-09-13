package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/controllers/members"
	"codeberg.org/mjh/LibRate/models"
)

func SetupProfiles(
	logger *zerolog.Logger,
	conf *cfg.Config,
	dbConn *sqlx.DB,
	app *fiber.App,
	fzlog *fiber.Handler,
) error {
	if err := setupStatic(app); err != nil {
		return err
	}

	api := app.Group("/api", *fzlog)

	mStor := models.NewMemberStorage(dbConn, logger, conf)
	memberSvc := members.NewController(mStor, logger)
	nicknames := mStor.GetNicknames()
	for i := range nicknames {
		app.Get("/"+nicknames[i], func(c *fiber.Ctx) error {
			return c.SendFile("./fe/build/profiles.html")
		})
	}
	app.Get("_app/*", func(c *fiber.Ctx) error {
		return c.SendFile("./fe/build/_app/" + c.Params("*"))
	})
	app.Get("*.css", func(c *fiber.Ctx) error {
		return c.SendFile("./fe/build/client/" + c.Params("*.css"))
	})
	app.Get("/:nickname", func(c *fiber.Ctx) error {
		return c.SendFile("./fe/build/profiles.html")
	})
	app.Get("/*/index.html", func(c *fiber.Ctx) error {
		return c.SendFile("./fe/build/profiles.html")
	})
	app.Get("/*/*.css", func(c *fiber.Ctx) error {
		return c.SendFile("./fe/build/" + c.Params("*.css"))
	})
	app.Get("/*/_app", func(c *fiber.Ctx) error {
		return c.SendFile("./fe/build/_app/" + c.Params("*"))
	})

	member := api.Group("/members")
	member.Get("/:id", memberSvc.GetMember)
	member.Get("/:nickname/info", memberSvc.GetMemberByNick)

	return nil
}
