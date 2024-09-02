package models

import (
	"time"
)

// Модель пользователя.
type User struct {
	ID               string    `json:"id"`
	UserName         string    `json:"user_name"`
	HashPassword     string    `json:"-"`
	Deleted          bool      `json:"-"`
	RegistrationDate time.Time `json:"registration_date"`
}
