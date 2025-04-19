package models

type Video struct {
	ID      int64
	FileID  string
	Caption string
	Tags    []string
}

type Tag struct {
	ID   int64
	Name string
}
