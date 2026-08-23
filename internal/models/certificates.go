package models

import "time"

type SignedCertificate struct {
	NodeName    string    `json:"node_name"`
	SignedAt    time.Time `json:"signed_at"`
	Certificate string    `json:"certificate"`
}

type RevokedCertificate struct {
	NodeName     string    `json:"node_name"`
	SerialNumber string    `json:"serial_number"`
	RevokedAt    time.Time `json:"revoked_at"`
}
