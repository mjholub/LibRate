package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/django/v3"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"

	"codeberg.org/mjh/LibRate/controllers/members"
	"codeberg.org/mjh/LibRate/models"
	"codeberg.org/mjh/LibRate/routes"
)

func setupNoscript() (*fiber.App, error) {
	engine := django.New("./views", ".django")
	noscript := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			err = c.Status(code).SendFile(fmt.Sprintf("./views/%d.html", code))
			if err != nil {
				return c.Status(500).SendString("Internal Server Error")
			}
			return nil
		},
		Views: engine,
	})
	// noscript.Use(fzlog)
	if err := routes.SetupNoScript(noscript); err != nil {
		return nil, err
	}
	return noscript, nil
}

func routeNoScript(app *fiber.App,
	db *sqlx.DB,
	log *zerolog.Logger,
	conf *cfg.Config,
) error {
	latestTag, err := getLatestTag()
	if err != nil {
		latestTag = "unknown"
		return fmt.Errorf("failed to get latest tag: %w", err)
	}
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"version": latestTag,
		})
	})

	mStorage := models.NewMemberStorage(db, log, conf)
	memberCon := members.NewController(mStorage, log)

	app.Get("/profiles/:nick", func(c *fiber.Ctx) error {
		member := memberCon.GetMemberByNick(c)
		return c.Render("profiles/member_page.django", fiber.Map{
			"member": member,
		})
	})

	return nil
}

func getLatestTag() (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	latestTag := strings.TrimSpace(string(out))
	return latestTag, nil
}
