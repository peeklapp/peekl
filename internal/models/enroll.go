package models

import "time"

type EnrollmentToken struct {
	TokenHash  string    `json:"token_hash"`
	Ip         string    `json:"ip"`
	ValidUntil time.Time `json:"valid_until"`
	CreatedAt  time.Time `json:"created_at"`
}
