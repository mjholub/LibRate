package models

type Film struct {
	Cast Cast
	Year uint32 `json:"year"`
}
