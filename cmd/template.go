package cmd

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/middleware/render"
)

func SetupTemplatedPages(conf *cfg.Config, app *fiber.App) {
	pages, err := render.MarkdownToHTML(conf.Fiber.StaticDir)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to render pages from markdown")
	}

	languages := lo.Uniq(lo.Map(pages, func(entry render.HTMLPage, index int) string {
		return strings.Split(strings.Split(entry.Name, "_")[1], ".")[0]
	}))
	log.Debug().Msgf("Languages: %+v", languages)
	fileNames := lo.Uniq(lo.Map(pages, func(entry render.HTMLPage, index int) string {
		return strings.Split(entry.Name, "_")[0]
	}))
	log.Debug().Msgf("File names: %+v", fileNames)

	for i := range fileNames {
		currentFileName := fileNames[i]
		app.Get("/"+currentFileName+"*", func(c *fiber.Ctx) error {
			path := strings.Split(c.Path(), "/")
			requestedDoc := path[len(path)-1]
			langName := strings.Split(strings.Split(requestedDoc, "_")[1], ".")[0]
			if !lo.Contains(languages, langName) {
				// redirect to default language
				c.Set("Content-Type", "text/html")
				page, ok := lo.Find(pages, func(entry render.HTMLPage) bool {
					return strings.Contains(entry.Name, currentFileName+"_"+conf.Fiber.DefaultLanguage)
				})
				if !ok {
					return c.Send(pages[0].Data)
				}
				return c.Send(page.Data)
			}
			for j := range pages {
				currentPage := pages[j]
				if strings.HasPrefix(currentPage.Name, currentFileName+"_") {
					c.Set("Content-Type", "text/html")
					return c.Send(currentPage.Data)
				}
			}
			return c.SendStatus(404)
		})
	}
}
