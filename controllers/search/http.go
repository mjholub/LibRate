package search

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"

	"codeberg.org/mjh/LibRate/controllers/search/aggregation"
	"codeberg.org/mjh/LibRate/controllers/search/target"
	h "codeberg.org/mjh/LibRate/internal/handlers"
)

// TODO: add aggregations documentation
// TODO: make better use of POST support
// @Summary Perform a search for the given query and options
// @Description Search for media, users, posts, artists, etc.
// @Tags search,media,metadata,users,posts,reviews
// @Param X-CSRF-Token header string false "CSRF token. Required when using POST."
// @Param q query string false "The search query. Falls back to a wildcard query if not provided."
// @Param category query string false "The category to search in" Enums(union,users,groups,artists,media,posts,reviews,tags,genres)
// @Param aggregations query string false "The aggregations to perform on the search results"
// @Param fuzzy query boolean false "Whether to perform a fuzzy search"
// @Param sort query string false "The field to sort the results by" Enums(score,added,modified,weighed_score,review_count)
// @Param desc query boolean false "Whether to sort the results in descending order"
// @Param page query integer false "The page to return"
// @Param pageSize query integer false "The number of results to return per page"
// @Accept json
// @Router /search [post]
// @Router /search [get]
func (s *Service) HandleSearch(c *fiber.Ctx) error {
	opts, err := parseQueries(c, s.validation)
	if err != nil {
		clientErr := strings.TrimSuffix(err.Error(), ":")
		return h.BadRequest(s.log, c, clientErr, "invalid input for "+c.Query("term"), err)
	}
	s.log.Debug().Msg("parsed queries")
	results, err := s.RunQuery(c.Context(), opts)
	if err != nil {
		return h.InternalError(s.log, c, "search failed", err)
	}

	return c.JSON(results)
}

func parseQueries(c *fiber.Ctx, v *validator.Validate) (opts *Options, err error) {
	// if no search term is provided, we'll create a "*" wildcard query
	searchTerm := c.Query("q", "")
	opts.Query = searchTerm

	category := c.Query("category", "union")
	if !target.ValidateCategory(category) {
		return nil, fmt.Errorf("invalid category: %q", category)
	}
	categoriesList := strings.Split(category, ",")
	categories := lo.Map(categoriesList, func(c string, _ int) target.Category {
		return target.FromStr(c)
	})
	opts.Categories = categories

	aggregations := c.Query("aggregations", "")
	if aggregations != "" {
		aggregationsList := strings.Split(aggregations, ",")
		if err := validateAggregations(target.FromStr(category), aggregationsList); err != nil {
			return nil, err
		}
		aggs := aggregation.FromStringSlice(aggregationsList)
		opts.Aggregations = aggs
	}

	fuzzy := c.QueryBool("fuzzy", false)
	opts.Fuzzy = fuzzy

	sort := c.Query("sort", "added")
	opts.Sort = sort
	if err := v.StructPartialCtx(c.Context(), opts, "sort"); err != nil {
		return nil, fmt.Errorf("invalid sort: %v", err)
	}

	sortDescending := c.QueryBool("desc", true)
	opts.SortDescending = sortDescending

	page := uint(c.QueryInt("page", 0))
	opts.Page = page

	pageSize := uint(c.QueryInt("pageSize", 0))
	opts.PageSize = pageSize
	if err := v.StructPartialCtx(c.Context(), opts, "pageSize"); err != nil {
		return nil, fmt.Errorf("invalid page size: %v", err)
	}

	return opts, nil
}

// check if an aggregation is possible for the given category
func validateAggregations(category target.Category, agg []string) error {
	switch category {
	case target.Union:
		return nil
	case target.Users, target.Groups:
		userAggregations := lo.Map(aggregation.UserAggregations, func(a aggregation.UserAggregation, _ int) string {
			return a.String()
		})
		if lo.Every(userAggregations, agg) {
			return nil
		}
		return fmt.Errorf("invalid aggregation for category %q: %v", category, agg)
	case target.Artists:
		artistAggregations := lo.Map(aggregation.ArtistAggregations, func(a aggregation.ArtistAggregation, _ int) string {
			return a.String()
		})
		if lo.Every(artistAggregations, agg) {
			return nil
		}
		return aggErr(category, agg)
	case target.Media:
		mediaAggregations := lo.Map(aggregation.MediaAggregations, func(a aggregation.MediaAggregation, _ int) string {
			return a.String()
		})
		if lo.Every(mediaAggregations, agg) {
			return nil
		}
		return aggErr(category, agg)
	case target.Posts, target.Reviews, target.Tags:
		postAggregations := lo.Map(aggregation.PostAggregations, func(a aggregation.PostAggregation, _ int) string {
			return a.String()
		})
		if lo.Every(postAggregations, agg) {
			return nil
		}
		return aggErr(category, agg)
	default:
		return fmt.Errorf("aggregations not supported for category %q", category)
	}
}

func aggErr(category target.Category, agg []string) error {
	return fmt.Errorf("invalid aggregation for category %q: %v", category, agg)
}
