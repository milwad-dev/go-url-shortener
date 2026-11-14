package domain

import "time"

type URL struct {
	ID          string    `json:"id"`
	OriginalURL string    `json:"original_url"`
	Code        string    `json:"code"`
	CreatedAt   time.Time `json:"created_at"`
}
