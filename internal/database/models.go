package database

import "time"

type ShortUrlModel struct {
	Id             int
	Link           string
	TimesClicked   int
	ExpTimeMinutes int
	CreatedAt      time.Time
	ShortCode      string
}

