package render

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/text/language"

	"github.com/goccy/go-yaml"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/storage/redis/v3"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/rs/zerolog"

	"github.com/fsnotify/fsnotify"
)

type HTMLPage struct {
	Name string `json:"name"`
	Data []byte `json:"data"`
}

func MarkdownToHTML(path string) (page *HTMLPage, err error) {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.LaxHTMLBlocks
	p := parser.NewWithExtensions(extensions)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank | html.CompletePage

	title := localizePageTitle(path)
	opts := html.RendererOptions{
		Flags: htmlFlags,
		Title: title,
		Icon:  "favicon.png",
	}
	renderer := html.NewRenderer(opts)

	baseName := filepath.Base(path)

	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", path, err)
	}
	return &HTMLPage{
		Name: strings.Replace(baseName, ".md", ".html", 1), // e.g. "en.html"
		Data: markdown.ToHTML(contents, p, renderer),
	}, nil
}

func localizePageTitle(path string) string {
	bundle := i18n.NewBundle(language.AmericanEnglish)

	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	// tos or privacy
	fileName := strings.Split(filepath.Base(path), ".")[0]

	// ISO 639-1 language code
	lang := strings.Split(fileName, "_")[1]

	_, err := bundle.LoadMessageFile(filepath.Join(filepath.Dir(path), "i18n", lang+".all.yml"))
	if err != nil {
		if strings.HasPrefix(fileName, "tos") {
			return "Terms of Service – LibRate"
		} else {
			return "Privacy Policy - LibRate"
		}
	}

	localizer := i18n.NewLocalizer(bundle, lang)

	if strings.HasPrefix(fileName, "tos") {
		return localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "tos",
				Other: "Terms of Service – LibRate",
			},
		})
	} else {
		return localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "privacy",
				Other: "Privacy Policy - LibRate",
			},
		})
	}
}

func preload(staticDir, target string) ([]HTMLPage, error) {
	files, err := os.ReadDir(staticDir)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %w", err)
	}

	var pages []HTMLPage

	for _, currentFile := range files {
		if currentFile.IsDir() {
			continue
		}

		if !strings.HasSuffix(currentFile.Name(), ".md") {
			continue
		}

		if strings.HasPrefix(currentFile.Name(), target) {
			page, err := MarkdownToHTML(filepath.Join(staticDir, currentFile.Name()))
			if err != nil {
				return nil, fmt.Errorf("error converting markdown to html: %w", err)
			}
			pages = append(pages, *page)
		}
	}

	return pages, nil
}

func WatchFiles(ctx context.Context, staticDir string, log *zerolog.Logger, cache *redis.Storage) error {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		privacyPages, err := preload(filepath.Join(staticDir, "templates"), "privacy")
		if err != nil {
			log.Error().Err(err).Msg("error preloading privacy policy pages")
		}
		for i := range privacyPages {
			if err = cache.Set(privacyPages[i].Name, privacyPages[i].Data, 0); err != nil {
				log.Error().Err(err).Msg("error setting privacy policy pages in cache")
			}
		}
	}()
	go func() {
		defer wg.Done()

		tosPages, err := preload(filepath.Join(staticDir, "templates"), "tos")
		if err != nil {
			log.Error().Err(err).Msg("error preloading terms of service pages")
		}
		for i := range tosPages {
			if err = cache.Set(tosPages[i].Name, tosPages[i].Data, 0); err != nil {
				log.Error().Err(err).Msg("error setting terms of service pages in cache")
			}
		}
	}()
	wg.Wait()

	log.Debug().Msg("Written initial pages to cache")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating watcher: %w", err)
	}
	defer watcher.Close()

	err = watcher.Add(filepath.Join(staticDir, "templates"))
	if err != nil {
		return fmt.Errorf("error adding directory to watcher: %w", err)
	}
	log.Info().Msg("Watching for changes in templates directory")

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("watcher events channel closed")
			}
			// FIXME: does not properly ignore vim temporary files with trailing ~
			if event.Op&fsnotify.Write == fsnotify.Write {
				// ignore temporary files
				fileName := filepath.Base(event.Name)
				if strings.HasSuffix(fileName, "~") {
					continue
				}
				fmt.Println("modified file:", event.Name)
				page, err := MarkdownToHTML(event.Name)
				if err != nil {
					log.Error().Err(err).Msg("error converting markdown to html")
				}
				if err := cache.Delete(strings.Replace(filepath.Base(event.Name), ".md", ".html", 1)); err != nil {
					log.Error().Err(err).Msg("error deleting privacy policy pages from cache")
				}

				if err := cache.Set(page.Name, page.Data, 0); err != nil {
					log.Error().Err(err).Msg("error setting privacy policy pages in cache")
				}

			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("watcher errors channel closed")
			}
			fmt.Println("error:", err)
		}
	}
}

func SetupTemplatedPages(
	defaultLang string,
	app *fiber.App,
	log *zerolog.Logger,
	cache *redis.Storage,
) {
	app.Get("/privacy/:language", func(c *fiber.Ctx) error {
		lang := c.Params("language")
		if err := loadPage(c, log, cache, "privacy", lang, defaultLang); err != nil {
			log.Error().Err(err).Msg("error loading privacy policy page")
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		// status 200 is carried within the loadPage function
		return nil
	})

	app.Get("/tos/:language", func(c *fiber.Ctx) error {
		lang := c.Params("language")
		if err := loadPage(c, log, cache, "tos", lang, defaultLang); err != nil {
			log.Error().Err(err).Msg("error loading terms of service page")
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		return nil
	})
}

func loadPage(c *fiber.Ctx, log *zerolog.Logger, cache *redis.Storage, target, lang, defaultLang string) error {
	c.Set("Content-Type", "text/html")
	pageData, err := cache.Get(target + "_" + lang + ".html")
	// fiber redis adapter returns nil, nil for redis.Nil
	if pageData == nil {
		fallback, err := cache.Get(target + "_" + defaultLang + ".html")
		if err != nil {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return c.Send(fallback)
	}

	if err != nil {
		return fmt.Errorf("error getting page from cache: %w", err)
	}

	return c.Send(pageData)
}
