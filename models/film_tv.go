package models

import "time"

type Film struct {
	ID    int    `json:"id" db:"id"`
	Title string `json:"title" db:"title"`
	Cast  Cast   `json:"cast" db:"cast"`
	Year  int    `json:"year" db:"year"`
}

func (f Film) GetID() int {
	return f.ID
}

func (f Film) IsMedia() bool {
	return true
}

type TVShow struct {
	ID      int      `json:"id" db:"id"`
	Title   string   `json:"title" db:"title"`
	Cast    Cast     `json:"cast" db:"cast"`
	Year    int      `json:"year" db:"year"`
	Active  bool     `json:"active" db:"active"`
	Seasons []Season `json:"seasons" db:"seasons"`
	Studio  Studio   `json:"studio" db:"studio"`
}

type Season struct {
	ID       int       `json:"id" db:"id"`
	ShowID   int       `json:"show_id" db:"show_id"`
	Number   uint16    `json:"number" db:"number"`
	Episodes []Episode `json:"episodes" db:"episodes"`
}

type Episode struct {
	ID        int           `json:"id" db:"id"`
	Title     string        `json:"title" db:"title"`
	Season    uint16        `json:"season" db:"season"`
	Episode   uint16        `json:"episode" db:"episode"`
	AirDate   time.Time     `json:"air_date" db:"air_date"`
	Duration  time.Duration `json:"duration" db:"duration"`
	Languages []string      `json:"languages" db:"languages"`
	Plot      string        `json:"plot" db:"plot"`
}

type Cast struct {
	Actors    []Person `json:"actors"`
	Directors []Person `json:"directors"`
}
