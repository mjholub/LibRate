package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/django/v3"
	"github.com/jmoiron/sqlx"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/models/member"

	"codeberg.org/mjh/LibRate/controllers/members"
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
	if err := routes.SetupNoScript(noscript); err != nil {
		return nil, err
	}
	return noscript, nil
}

func routeNoScript(app *fiber.App,
	sqlDriver *sqlx.DB,
	log *zerolog.Logger,
	conf *cfg.Config,
	neo4jDriver neo4j.DriverWithContext,
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

	// setup member storage
	var mStorage member.MemberStorer
	switch conf.Engine {
	// note that only postgres is actually supported currently
	case "postgres", "mariadb", "sqlite":
		mStorage = member.NewSQLStorage(sqlDriver, log, conf)
	case "neo4j":
		mStorage = member.NewNeo4jStorage(neo4jDriver, log, conf)
	default:
		return fmt.Errorf("unsupported database engine: %s", conf.Engine)
	}

	memberCon := members.NewController(mStorage, log, conf)

	app.Get("/profiles/:nick", func(c *fiber.Ctx) error {
		member := memberCon.GetMemberByNick(c)
		return c.Render("profiles/member_page.django", fiber.Map{
			"member": member,
		})
	})

	return nil
}

func getLatestTag() (string, error) {
	// check if git is present, otherwise try os.GetEnv("GIT_TAG")
	if _, err := exec.LookPath("git"); err != nil {
		if tag := strings.TrimSpace(os.Getenv("GIT_TAG")); tag != "" {
			return tag, nil
		}
		return "", err
	}
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	latestTag := strings.TrimSpace(string(out))
	return latestTag, nil
}
