package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-json"
)

type GenreInfo struct {
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	Subgenres   []GenreInfo `json:"children,omitempty"`
}

func main() {
	// list all json files in the current directory
	jsonFiles, err := filepath.Glob("data/*.json")
	if err != nil {
		panic(fmt.Errorf("error listing json files: %v", err))
	}

	var queries []string
	for _, file := range jsonFiles {
		file = strings.Split(file, ".")[0]
		base := strings.Split(file, "/")[1]
		title := strings.Title(base)
		query := fmt.Sprintf(`
		INSERT INTO media.genres (name, kinds) VALUES '%s', ARRAY ['music'] ON CONFLICT DO NOTHING
		RETURNING id;`, title)
		queries = append(queries, query)

		f, err := os.ReadFile(file + ".json")
		if err != nil {
			panic(fmt.Errorf("error reading file %s: %v", file, err))
		}
		genres := []GenreInfo{}
		err = json.Unmarshal(f, &genres)
		if err != nil {
			panic(fmt.Errorf("error unmarshalling file %s: %v", file, err))
		}
		for _, g := range genres {
			secondaryGenreQuery := fmt.Sprintf(`INSERT INTO media.genres (name, kinds, parent)
			VALUES %s, ARRAY ['music'], (SELECT id FROM media.genres WHERE name = '%s')
			ON CONFLICT DO NOTHING;`, g.Name, title)

			descriptionQuery := fmt.Sprintf(`INSERT INTO media.genre_descriptions (language, description, genre_id)
			VALUES ('en', '%s', (SELECT id FROM media.genres WHERE name = '%s'));`, g.Description, g.Name)
			queries = append(queries, secondaryGenreQuery)
			queries = append(queries, descriptionQuery)

			for _, subgenre := range g.Subgenres {
				query := fmt.Sprintf(
					`INSERT INTO media.genres (name, kinds, parent)
				VALUES (%s, ARRAY ['music'], (SELECT id FROM media.genres WHERE name = %s))
				`, subgenre.Name, g.Name)
				q2 := fmt.Sprintf(`INSERT INTO media.genre_descriptions (language, description, genre_id)
				VALUES ('en', '%s', (SELECT id FROM media.genres WHERE name = '%s'));`, subgenre.Description, subgenre.Name)
				queries = append(queries, query)
				queries = append(queries, q2)
			}
			// then add children to the parent
			query := fmt.Sprintf(`UPDATE media.genres
			SET children = ARRAY(SELECT id FROM media.genres 
			WHERE parent = (SELECT id FROM media.genres WHERE name = '%s')) 
			WHERE name = '%s';`, title, title)
			queries = append(queries, query)
		}

		err = os.WriteFile("queries.sql", []byte(strings.Join(queries, "\n")), 0o644)
		if err != nil {
			panic(fmt.Errorf("error writing queries to file: %v", err))
		}
	}
}
