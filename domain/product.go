package domain

import "time"

type Product struct {
	ID           int64      `db:"id"`
	Name         string     `db:"name"`
	Image        string     `db:"image"`
	Category     string     `db:"category"`
	Description  string     `db:"description"`
	Rating       int64      `db:"rating"`
	NumReviews   int64      `db:"num_reviews"`
	Price        float64    `db:"price"`
	CountInStock int64      `db:"count_in_stock"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    *time.Time `db:"updated_at"`
}
