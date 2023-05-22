package models

type Film struct {
	Cast Cast
	Year uint32 `json:"year"`
}

type Cast struct {
	Actors    []Person `json:"actors"`
	Directors []Person `json:"directors"`
}
