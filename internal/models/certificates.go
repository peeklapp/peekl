package models

import "time"

type PendingCertificate struct {
	NodeName    string    `json:"node_name"`
	SubmittedAt time.Time `json:"submitted_at"`
	Data        string    `json:"data"`
}

type SignedCertificate struct {
	NodeName     string    `json:"node_name"`
	CsrSignature string    `json:"csr_signature"`
	SignedAt     time.Time `json:"signed_at"`
	Data         string    `json:"data"`
}

type RevokedCertificate struct {
	NodeName     string    `json:"node_name"`
	SerialNumber string    `json:"serial_number"`
	RevokedAt    time.Time `json:"revoked_at"`
}
