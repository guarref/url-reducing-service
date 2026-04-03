package model

import "time"

type Link struct {
	ID          int64     `db:"id" json:"id"`
	OriginalURL string    `db:"original_url" json:"original_url"`
	ShortCode   string    `db:"short_code" json:"short_code"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}
