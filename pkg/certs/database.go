package certs

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// This module contains all the code related to the database.
//
// The database is used to store information about what the
// PKI has signed, especially the signature of the CSR that
// has been signed, so that it can be used as a unique
// identifier when the agent come back to get it's signed
// certificate.

// Error when CSR signature is not found
type PendingCertificateNotFound struct {
	nodeName string
}

func (c PendingCertificateNotFound) Error() string {
	return fmt.Sprintf("No pending certificate found for node %s", c.nodeName)
}

type CertsDatabaseEngine struct {
	DB *sql.DB
}

// PENDING CERTIFICATES
type PendingCertificate struct {
	NodeName    string    `json:"node_name"`
	Signature   string    `json:"signature,omitempty"`
	SubmittedAt time.Time `json:"submitted_at"`
}

func (c *CertsDatabaseEngine) InsertPendingCertificate(nodeName string, csrSignature string) error {
	_, err := c.DB.Exec("INSERT INTO pending_certs (node_name, signature) VALUES (?, ?)", nodeName, csrSignature)
	if err != nil {
		return fmt.Errorf("Could not insert pending certificate in database : %s", err.Error())
	}
	return nil
}

func (c *CertsDatabaseEngine) GetPendingCertificate(nodeName string) (PendingCertificate, error) {
	var pendingCert PendingCertificate

	err := c.DB.QueryRow(
		"SELECT node_name, signature, submitted_at FROM pending_certs WHERE node_name = ?", nodeName,
	).Scan(&pendingCert.NodeName, &pendingCert.Signature, &pendingCert.SubmittedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return pendingCert, PendingCertificateNotFound{nodeName: nodeName}
		} else {
			return pendingCert, err
		}
	}

	return pendingCert, nil
}

func (c *CertsDatabaseEngine) ListPendingCertificates() ([]PendingCertificate, error) {
	var pendCerts []PendingCertificate

	rows, err := c.DB.Query("SELECT node_name, submitted_at FROM pending_certs")
	if err != nil {
		return pendCerts, err
	}
	defer rows.Close()

	for rows.Next() {
		var s PendingCertificate
		err := rows.Scan(&s.NodeName, &s.SubmittedAt)
		if err != nil {
			return pendCerts, err
		}
		pendCerts = append(pendCerts, s)
	}

	return pendCerts, nil
}

func (c *CertsDatabaseEngine) DeletePendingCertificate(nodeName string) error {
	_, err := c.DB.Exec("DELETE FROM pending_certs WHERE node_name = ?", nodeName)
	if err != nil {
		return err
	}
	return nil
}

// SIGNED CERTIFICATES

// Error when CSR signature is not found
type SignedCertificateNotFound struct {
	nodeName string
}

func (c SignedCertificateNotFound) Error() string {
	return fmt.Sprintf("No signed certificate found for node %s", c.nodeName)
}

type SignedCertificate struct {
	NodeName     string    `json:"node_name"`
	CsrSignature string    `json:"csr_signature"`
	SignedAt     time.Time `json:"signed_at"`
}

func (c *CertsDatabaseEngine) InsertSignedCertificate(nodeName string, csrSignature string) error {
	_, err := c.DB.Exec(
		"INSERT INTO signed_certs (node_name, csr_signature) VALUES (?, ?)", nodeName, csrSignature,
	)
	if err != nil {
		return fmt.Errorf("Could not insert signed certificate in database : %s", err.Error())
	}
	return nil
}

func (c *CertsDatabaseEngine) GetSignedCertificate(nodeName string) (SignedCertificate, error) {
	var signedCert SignedCertificate
	err := c.DB.QueryRow(
		"SELECT node_name, csr_signature, signed_at FROM signed_certs WHERE node_name = ?", nodeName,
	).Scan(&signedCert.NodeName, &signedCert.CsrSignature, &signedCert.SignedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return signedCert, SignedCertificateNotFound{nodeName: nodeName}
		} else {
			return signedCert, err
		}
	}
	return signedCert, nil
}

func (c *CertsDatabaseEngine) ListSignedCertificates() ([]SignedCertificate, error) {
	var signedCerts []SignedCertificate

	rows, err := c.DB.Query("SELECT node_name, signed_at FROM signed_certs")
	if err != nil {
		return signedCerts, err
	}
	defer rows.Close()

	for rows.Next() {
		var s SignedCertificate
		err := rows.Scan(&s.NodeName, &s.SignedAt)
		if err != nil {
			return signedCerts, err
		}
		signedCerts = append(signedCerts, s)
	}

	return signedCerts, nil
}

func NewCertsDatabaseEngine(databasePath string) (*CertsDatabaseEngine, error) {
	var engine CertsDatabaseEngine

	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return &engine, err
	}

	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS pending_certs (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      node_name TEXT NOT NULL UNIQUE,
      signature TEXT NOT NULL UNIQUE,
      submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS signed_certs (
    	id INTEGER PRIMARY KEY AUTOINCREMENT,
    	node_name TEXT NOT NULL UNIQUE,
    	csr_signature TEXT NOT NULL UNIQUE,
    	signed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
	`)
	if err != nil {
		return &engine, err
	}

	engine.DB = db
	return &engine, nil
}
