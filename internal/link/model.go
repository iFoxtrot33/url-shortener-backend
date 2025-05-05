package link

import (
	"math/rand"
	"time"

	"gorm.io/gorm"
)

// Link represents a shortened link model
// @Description Shortened link model
type Link struct {
	ID             uint           `json:"id" gorm:"primaryKey" example:"1"`
	CreatedAt      time.Time      `json:"created_at" example:"2025-04-23T00:00:00Z"`
	UpdatedAt      time.Time      `json:"updated_at" example:"2025-04-23T00:00:00Z"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at,omitempty" swaggertype:"string" format:"date-time"`
	Url            string         `json:"url" example:"https://example.com"`
	Hash           string         `json:"hash" gorm:"index:,unique,where:deleted_at IS NULL" example:"abc123"`
	UserId         string         `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	NumberOfClicks int64          `json:"number_of_clicks" gorm:"default:0" example:"42"`
	Lifetime       int64          `json:"lifetime" example:"90"`
}

func NewLink(url string) *Link {
	return &Link{
		Url:  url,
		Hash: RandStringRunes(10),
	}
}

var letterRunes = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
