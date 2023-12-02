package routes

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/controllers/members"
	"codeberg.org/mjh/LibRate/models/member"
)

func SetupProfiles(
	logger *zerolog.Logger,
	conf *cfg.Config,
	dbConn *sqlx.DB,
	app *fiber.App,
) error {
	if err := setupStatic(app); err != nil {
		return err
	}

	api := app.Group("/api")

	var mStor member.MemberStorer

	switch conf.Engine {
	case "postgres", "sqlite", "mariadb":
		mStor = member.NewSQLStorage(dbConn, logger, conf)
	default:
		return fmt.Errorf("unsupported database engine \"%q\" or error reading config", conf.Engine)
	}
	memberSvc := members.NewController(mStor, dbConn, logger, conf)
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
	member.Get("/:email_or_username/info", memberSvc.GetMemberByNickOrEmail)

	return nil
}
