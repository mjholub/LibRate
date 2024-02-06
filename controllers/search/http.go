package search

import (
	"fmt"

	"codeberg.org/mjh/LibRate/controllers/search/target"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// @Summary Perform a search for the given query and options
// @Description Search for media, users, posts, artists, etc.
// @Tags search,media,metadata,users,posts,reviews
// @Param term query string true "The search term"
// @Accept json
// Router /search [post]
func (s *Service) HandleSearch(c *fiber.Ctx) error {
	searchTerm := c.Query("term")
}

func parseQueries(c *fiber.Ctx, v *validator.Validate) ([]string, []uint, []bool, error) {
	searchTerm := c.Query("term")
	if searchTerm == "" {
		return nil, nil, nil, fmt.Errorf("missing search term")
	}
	category := c.Query("category", "union")
	if !target.ValidateCategory(category) {
		return nil, nil, nil, fmt.Errorf("invalid category: %q", category)
	}
	sort := c.Query("sort", "added")
	o := &Options{
		Sort: sort,
	}
	if err := v.StructPartialCtx(c.Context(), o, "sort"); err != nil {
		return nil, nil, nil, fmt.Errorf("invalid sort: %v", err)
	}
	sortDescending := c.QueryBool("desc", true)
	page := uint(c.QueryInt("page", 0))
	pageSize := uint(c.QueryInt("pageSize", 0))
	o.PageSize = pageSize
	if err := v.StructPartialCtx(c.Context(), o, "pageSize"); err != nil {
		return nil, nil, nil, fmt.Errorf("invalid page size: %v", err)
	}

	return []string{searchTerm, category, sort}, []uint{page, pageSize}, []bool{sortDescending}, nil
}

func parseFilters(c *fiber.Ctx, v *validator.Validate)
