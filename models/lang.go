package models

import "fmt"

type LangID int16

const (
	English LangID = iota
	Spanish
	French
	German
	Chinese
	Japanese
	Korean
	Arabic
	Hebrew
	Hindi
	Polish
	Russian
	Czech
	Dutch
	Greek
	Italian
	Swedish
	Turkish
	Norwegian
	Portuguese
	Finnish
	Thai
	Indonesian
	Vietnamese
	Farsi
	Tagalog
	Swahili
	Serbian
	Croatian
	Bosnian
	Slovenian
	Slovak
	Macedonian
	Albanian
	Bulgarian
	Romanian
	Hungarian
	Latvian
	Lithuanian
	Estonian
	Ukrainian
	Belarusian
	Malay
	Malayalam
	Tamil
	Telugu
	Kannada
	Marathi
	Gujarati
	Bengali
	Punjabi
	Urdu
	Mongolian
	Amharic
	Icelandic
	Maltese
	Unknown
	Others
)

var langNameToID = map[string]LangID{
	"English":    English,
	"Spanish":    Spanish,
	"French":     French,
	"German":     German,
	"Chinese":    Chinese,
	"Japanese":   Japanese,
	"Korean":     Korean,
	"Arabic":     Arabic,
	"Hebrew":     Hebrew,
	"Hindi":      Hindi,
	"Polish":     Polish,
	"Russian":    Russian,
	"Czech":      Czech,
	"Dutch":      Dutch,
	"Greek":      Greek,
	"Italian":    Italian,
	"Swedish":    Swedish,
	"Turkish":    Turkish,
	"Norwegian":  Norwegian,
	"Portuguese": Portuguese,
	"Finnish":    Finnish,
	"Thai":       Thai,
	"Indonesian": Indonesian,
	"Vietnamese": Vietnamese,
	"Farsi":      Farsi,
	"Tagalog":    Tagalog,
	"Swahili":    Swahili,
	"Serbian":    Serbian,
	"Croatian":   Croatian,
	"Bosnian":    Bosnian,
	"Slovenian":  Slovenian,
	"Slovak":     Slovak,
	"Macedonian": Macedonian,
	"Albanian":   Albanian,
	"Bulgarian":  Bulgarian,
	"Romanian":   Romanian,
	"Hungarian":  Hungarian,
	"Latvian":    Latvian,
	"Lithuanian": Lithuanian,
	"Estonian":   Estonian,
	"Ukrainian":  Ukrainian,
	"Belarusian": Belarusian,
	"Malay":      Malay,
	"Malayalam":  Malayalam,
	"Tamil":      Tamil,
	"Telugu":     Telugu,
	"Kannada":    Kannada,
	"Marathi":    Marathi,
	"Gujarati":   Gujarati,
	"Bengali":    Bengali,
	"Punjabi":    Punjabi,
	"Urdu":       Urdu,
	"Mongolian":  Mongolian,
	"Amharic":    Amharic,
	"Icelandic":  Icelandic,
	"Maltese":    Maltese,
	"Unknown":    Unknown,
	"Others":     Others,
}

// ReverseLookupLangID takes a language name as input and returns the corresponding LangID.
// If the language name is not found, it returns an error.
func ReverseLookupLangID(langName string) (LangID, error) {
	if id, ok := langNameToID[langName]; ok {
		return id, nil
	}
	return 0, fmt.Errorf("Language name not found: %s", langName)
}
