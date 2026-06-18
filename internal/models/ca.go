package models

import "time"

type CertificateAuthorityParameters struct {
	NotBefore time.Time
	NotAfter  time.Time
}
