package render

import (
	"fmt"
	"os"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type HTMLPage struct {
	Name string
	Data []byte
}

func MarkdownToHTML(staticDir string) (htmlPages []HTMLPage, err error) {
	files, err := os.ReadDir(staticDir + "/templates")
	if err != nil {
		return nil, fmt.Errorf("error reading templates directory: %w", err)
	}

	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank | html.CompletePage
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	for i := range files {
		if files[i].IsDir() {
			continue
		}
		contents, err := os.ReadFile(staticDir + "/templates/" + files[i].Name())
		if err != nil {
			return nil, fmt.Errorf("error reading file %s: %w", files[i].Name(), err)
		}
		page := HTMLPage{
			Name: strings.Replace(files[i].Name(), ".md", ".html", 1),
			Data: markdown.ToHTML(contents, p, renderer),
		}

		htmlPages = append(htmlPages, page)
	}

	return htmlPages, nil
}
