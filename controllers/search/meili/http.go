package meili

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"

	"codeberg.org/mjh/LibRate/controllers/search/target"
	h "codeberg.org/mjh/LibRate/internal/handlers"
)

// @Summary Perform a search for the given query and options
// @Description Search for media, users, posts, artists, etc.
// @Tags search,media,metadata,users,posts,reviews
// @Param X-CSRF-Token header string false "CSRF token. Required when using POST."
// @Param q query string false "The search query. Falls back to a wildcard query if not provided."
// @Param category query string false "The category to search in" Enums(union,members,artists,media,ratings,genres)
// @Param fuzzy query boolean false "Whether to perform a fuzzy search"
// @Param sort query string false "The field to sort the results by" Enums(score,added,modified,name)
// @Param desc query boolean false "Whether to sort the results in descending order"
// @Param page query integer false "The page to return"
// @Param pageSize query integer false "The number of results to return per page"
// @Accept json
// @Router /search [post]
// @Router /search [get]
func (s *Service) HandleSearch(c *fiber.Ctx) error {
	opts, err := s.parseQueries(c)
	if err != nil {
		clientErr := strings.TrimSuffix(err.Error(), ":")
		return h.BadRequest(s.log, c, clientErr, "invalid input for "+c.Query("term"), err)
	}
	s.log.Debug().Msg("parsed queries")
	results, err := s.RunQuery(opts)
	if err != nil {
		return h.InternalError(s.log, c, "search failed", err)
	}

	return c.JSON(results)
}

func (s *Service) parseQueries(c *fiber.Ctx) (opts *Options, err error) {
	opts = &Options{}
	// if no search term is provided, we'll create a "*" wildcard query
	err = c.QueryParser(opts)
	if err != nil {
		return nil, fmt.Errorf("error parsing queries: %v", err)
	}

	category := c.Query("category", "union")
	if !target.ValidateCategory(category) {
		return nil, fmt.Errorf("invalid category: %q", category)
	}
	s.log.Trace().Msgf("parsed category: %s", category)
	categoriesList := strings.Split(category, ",")
	categories := lo.Map(categoriesList, func(c string, _ int) target.Category {
		return target.FromStr(c)
	})
	opts.Categories = categories
	s.log.Trace().Msgf("parsed categories: %v", categories)

	pageSize := uint(c.QueryInt("pageSize", 0))
	opts.PageSize = pageSize
	if err := s.validation.StructPartialCtx(c.Context(), opts, "pageSize"); err != nil {
		return nil, fmt.Errorf("invalid page size: %v", err)
	}
	s.log.Trace().Msgf("parsed query parameters into options: %+v", opts)

	return opts, nil
}
