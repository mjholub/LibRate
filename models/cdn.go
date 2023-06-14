package models

type Image struct {
	ID        uint   `json:"id" db:"id,pk,unique,autoinc"`
	Source    string `json:"source" db:"source"`
	Thumbnail string `json:"thumbnail" db:"thumbnail"`
}
