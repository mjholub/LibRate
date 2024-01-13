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
		title := strings.ReplaceAll(strings.Title(base), "'", "''")

		// Main genre query
		mainGenreQuery := fmt.Sprintf(`
			INSERT INTO media.genres (name, kinds)
			VALUES ('%s', ARRAY['music'])
			ON CONFLICT DO NOTHING;`, strings.ReplaceAll(title, "'", "''"))

		queries = append(queries, mainGenreQuery)

		// Read and parse JSON file
		f, err := os.ReadFile(file + ".json")
		if err != nil {
			panic(fmt.Errorf("error reading file %s: %v", file, err))
		}
		genres := []GenreInfo{}
		err = json.Unmarshal(f, &genres)
		if err != nil {
			panic(fmt.Errorf("error unmarshalling file %s: %v", file, err))
		}

		// Process genres and subgenres
		for _, g := range genres {
			g.Name = strings.ReplaceAll(g.Name, "'", "''")
			g.Description = strings.ReplaceAll(g.Description, "'", "''")
			// Secondary genre query
			secondaryGenreQuery := fmt.Sprintf(`
				INSERT INTO media.genres (name, kinds, parent)
				VALUES ('%s', ARRAY['music'], (SELECT id FROM media.genres WHERE name = '%s'))
				ON CONFLICT DO NOTHING;`, g.Name, title)

			// Description query
			descriptionQuery := fmt.Sprintf(`
				INSERT INTO media.genre_descriptions (language, description, genre_id)
				VALUES ('en', '%s', (SELECT id FROM media.genres WHERE name = '%s'))
				ON CONFLICT DO NOTHING;`, g.Description, g.Name)

			queries = append(queries, secondaryGenreQuery)
			queries = append(queries, descriptionQuery)

			// Subgenres
			for _, subgenre := range g.Subgenres {
				subgenre.Name = strings.ReplaceAll(subgenre.Name, "'", "''")
				subgenre.Description = strings.ReplaceAll(subgenre.Description, "'", "''")
				subgenreQuery := fmt.Sprintf(`
					INSERT INTO media.genres (name, kinds, parent)
					VALUES ('%s', ARRAY['music'], (SELECT id FROM media.genres WHERE name = '%s'))
					ON CONFLICT DO NOTHING;`, subgenre.Name, g.Name)

				subgenreDescriptionQuery := fmt.Sprintf(`
					INSERT INTO media.genre_descriptions (language, description, genre_id)
					VALUES ('en', '%s', (SELECT id FROM media.genres WHERE name = '%s'))
					ON CONFLICT DO NOTHING;`, subgenre.Description, subgenre.Name)

				queries = append(queries, subgenreQuery)
				queries = append(queries, subgenreDescriptionQuery)
			}

			// Update parent with children
			updateParentQuery := fmt.Sprintf(`
				UPDATE media.genres
				SET children = ARRAY(SELECT id FROM media.genres WHERE parent = (SELECT id FROM media.genres WHERE name = '%s'))
				WHERE name = '%s';`, title, title)

			queries = append(queries, updateParentQuery)
		}
	}

	// Write queries to file
	err = os.WriteFile("queries.sql", []byte(strings.Join(queries, "\n")), 0o644)
	if err != nil {
		panic(fmt.Errorf("error writing queries to file: %v", err))
	}
}
