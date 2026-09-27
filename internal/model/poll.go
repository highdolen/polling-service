package model

import "time"

type Poll struct {
	ID        int64
	Question  string
	Type      string
	Status    string
	StartsAt  time.Time
	EndsAt    time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
