package render

import (
	"html/template"

	"github.com/gofiber/template/html/v2"

	"codeberg.org/mjh/LibRate/cfg"
)

func Setup(conf *cfg.Config) *html.Engine {
	engine := html.New("./views", ".html")
	if conf.LibrateEnv == "development" {
		engine.Reload(true)
		if conf.Logging.Level == "trace" {
			engine.Debug(true)
		}
	}
	engine.AddFunc("unescape", func(s string) template.HTML {
		return template.HTML(s)
	})

	return engine
}
