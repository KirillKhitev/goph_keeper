package models

import (
	"time"
)

type Data struct {
	ID          string    `json:"id,omitempty"`
	Name        []byte    `json:"name,omitempty"`
	Type        string    `json:"type,omitempty"`
	UserID      string    `json:"user_id,omitempty"`
	Deleted     bool      `json:"deleted,omitempty"`
	Date        time.Time `json:"date,omitempty"`
	Body        []byte    `json:"body,omitempty"`
	Description []byte    `json:"description,omitempty"`
}

type LoginBody struct {
	Login    string
	Password string
}

type CreditCardBody struct {
	Ccn string
	Exp string
	CVV string
}
