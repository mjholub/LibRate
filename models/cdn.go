package models

type Image struct {
	ID        int64  `json:"id" db:"id,pk,unique,autoinc"`
	Source    string `json:"source" db:"source"`
	Thumbnail string `json:"thumbnail" db:"thumbnail"`
}
