package domain

import "github.com/google/uuid"

type URL struct {
	ShortURL    string
	OriginalURL string
	UserID      uuid.UUID
}
