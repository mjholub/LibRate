package search

import "github.com/blevesearch/bleve/v2/search/query"

type Filter interface {
	filterQuery(params ...any) query.Query
}

func (f *TermsFilter) filterQuery(params ...map[string]interface{}) query.Query {
	if len(params) == 0 {
		return nil
	}

	for i := range params {
		if _, ok := params[i]["language"]; ok {
			return filterByLanguage(params[i]["language"].(string))
		}
	}

	return nil
}

func filterByLanguage(language string) query.Query {
	return nil
}
