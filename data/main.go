package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/samber/lo"
)

func main() {
	qBytes, err := os.ReadFile(filepath.Join("/", "home", "user", "Projects", "LibRate", "data", "queries2.cypher"))
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(qBytes), "\n")
	genresB, err := os.ReadFile(filepath.Join("/", "home", "user", "Projects", "LibRate", "data", "genres.txt"))
	gen := strings.Split(string(genresB), "\n")

	re := regexp.MustCompile(`'([A-Za-z0-9,.\-()\"\\' ]*)'`)

	matched := matchLines(lines, gen, re)

	replaced := lo.Associate(gen, func(s string) (string, string) {
		lowered := strings.ToLower(s)
		return s, strings.ReplaceAll(lowered, " ", "_")
	})

	linesToReplace := lo.Keys(matched)
	sort.Ints(linesToReplace)

	processLines(linesToReplace, matched, replaced, lines)

	data := []byte(strings.Join(lines, "\n"))
	if err := os.WriteFile("/home/user/Projects/list.txt", data, 0o640); err != nil {
		panic(fmt.Errorf("error writing to file: %v", err))
	}
}

func matchLines(lines, genres []string, re *regexp.Regexp) (lineNumWithSubstr map[int][]string) {
	lineNumWithSubstr = make(map[int][]string)
	for i, line := range lines {
		for _, genre := range genres {
			if strings.Contains(line, genre) && !re.MatchString(line) && !lo.Contains(lineNumWithSubstr[i], genre) && genre != "" {
				lineNumWithSubstr[i] = append(lineNumWithSubstr[i], genre)
			}
		}
	}
	return lineNumWithSubstr
}

func processLines(linesToReplace []int, matched map[int][]string, replaced map[string]string, lines []string) {
	// defer wg.Done()
	for i := range linesToReplace {
		for _, matchingStr := range matched[linesToReplace[i]] {
			lines[linesToReplace[i]] = strings.ReplaceAll(lines[linesToReplace[i]], string(matchingStr), replaced[string(matchingStr)])
		}
	}
}
