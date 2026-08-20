package database

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/peeklapp/peekl/internal/config"
	"github.com/peeklapp/peekl/internal/models"
	_ "modernc.org/sqlite"
)

type DatabaseEngine struct {
	DB     *sql.DB
	DBType string
}

// Pending certificates
func (c *DatabaseEngine) InsertPendingCertificate(nodeName string, csrData string) error {
	queries := map[string]string{
		"sqlite":   "INSERT INTO pending_certs (node_name, data) VALUES (?, ?);",
		"postgres": "INSERT INTO pending_certs (node_name, data) VALUES ($1, $2);",
	}

	_, err := c.DB.Exec(queries[c.DBType], nodeName, csrData)
	if err != nil {
		return fmt.Errorf("could not insert pending certificate in database : %s", err.Error())
	}
	return nil
}

func (c *DatabaseEngine) GetPendingCertificate(nodeName string) (models.PendingCertificate, error) {
	queries := map[string]string{
		"sqlite":   "SELECT node_name, submitted_at, data FROM pending_certs WHERE node_name = ?;",
		"postgres": "SELECT node_name, submitted_at, data FROM pending_certs WHERE node_name = $1;",
	}

	var pendingCert models.PendingCertificate

	err := c.DB.QueryRow(
		queries[c.DBType], nodeName,
	).Scan(&pendingCert.NodeName, &pendingCert.SubmittedAt, &pendingCert.Data)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return pendingCert, models.PendingCertificateNotFound{NodeName: nodeName}
		} else {
			return pendingCert, err
		}
	}

	return pendingCert, nil
}

func (c *DatabaseEngine) ListPendingCertificates() ([]models.PendingCertificate, error) {
	var pendCerts []models.PendingCertificate

	rows, err := c.DB.Query("SELECT node_name, submitted_at FROM pending_certs;")
	if err != nil {
		return pendCerts, err
	}
	defer rows.Close() //nolint:errcheck

	for rows.Next() {
		var s models.PendingCertificate
		err := rows.Scan(&s.NodeName, &s.SubmittedAt)
		if err != nil {
			return pendCerts, err
		}
		pendCerts = append(pendCerts, s)
	}

	return pendCerts, nil
}

func (c *DatabaseEngine) DeletePendingCertificate(nodeName string) error {
	queries := map[string]string{
		"sqlite":   "DELETE FROM pending_certs WHERE node_name = ?;",
		"postgres": "DELETE FROM pending_certs WHERE node_name = $1;",
	}
	_, err := c.DB.Exec(queries[c.DBType], nodeName)
	if err != nil {
		return err
	}
	return nil
}

// Signed certificates
func (c *DatabaseEngine) InsertSignedCertificate(nodeName string, csrSignature string, certificate string) error {
	queries := map[string]string{
		"sqlite":   "INSERT INTO signed_certs (node_name, csr_signature, data) VALUES (?, ?, ?);",
		"postgres": "INSERT INTO signed_certs (node_name, csr_signature, data) VALUES ($1, $2, $3);",
	}
	_, err := c.DB.Exec(queries[c.DBType], nodeName, csrSignature, certificate)
	if err != nil {
		return fmt.Errorf("could not insert signed certificate in database : %s", err.Error())
	}
	return nil
}

func (c *DatabaseEngine) GetSignedCertificate(csrSignature string) (models.SignedCertificate, error) {
	queries := map[string]string{
		"sqlite":   "SELECT node_name, csr_signature, signed_at, data FROM signed_certs WHERE csr_signature = ?;",
		"postgres": "SELECT node_name, csr_signature, signed_at, data FROM signed_certs WHERE csr_signature = $1;",
	}
	var signedCert models.SignedCertificate
	err := c.DB.QueryRow(
		queries[c.DBType], csrSignature,
	).Scan(&signedCert.NodeName, &signedCert.CsrSignature, &signedCert.SignedAt, &signedCert.Data)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return signedCert, models.SignedCertificateNotFound{CsrSignature: csrSignature}
		} else {
			return signedCert, err
		}
	}
	return signedCert, nil
}

func (c *DatabaseEngine) ListSignedCertificates() ([]models.SignedCertificate, error) {
	var signedCerts []models.SignedCertificate

	rows, err := c.DB.Query("SELECT node_name, signed_at FROM signed_certs")
	if err != nil {
		return signedCerts, err
	}
	defer rows.Close() //nolint:errcheck

	for rows.Next() {
		var s models.SignedCertificate
		err := rows.Scan(&s.NodeName, &s.SignedAt)
		if err != nil {
			return signedCerts, err
		}
		signedCerts = append(signedCerts, s)
	}

	return signedCerts, nil
}

func (c *DatabaseEngine) IsNodeNameUsed(nodeName string) (bool, error) {
	queries := map[string]string{
		"sqlite":   "SELECT node_name FROM signed_certs WHERE node_name = ?;",
		"postgres": "SELECT node_name FROM signed_certs WHERE node_name = $1;",
	}
	err := c.DB.QueryRow(queries[c.DBType], nodeName).Scan(&nodeName)
	if err != nil {
		if err != sql.ErrNoRows {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

func databaseSchema(dbType string) string {
	switch dbType {
	case "sqlite":
		return `
    CREATE TABLE IF NOT EXISTS pending_certs (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      node_name TEXT NOT NULL UNIQUE,
      submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    	data TEXT NOT NULL UNIQUE
    );

    CREATE TABLE IF NOT EXISTS signed_certs (
    	id INTEGER PRIMARY KEY AUTOINCREMENT,
    	node_name TEXT NOT NULL UNIQUE,
    	csr_signature TEXT NOT NULL UNIQUE,
    	signed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    	data TEXT NOT NULL UNIQUE
    );
		`
	case "postgres":
		return `
		CREATE TABLE IF NOT EXISTS pending_certs (
			id SERIAL PRIMARY KEY,
			node_name TEXT NOT NULL UNIQUE,
			submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			data TEXT NOT NULL UNIQUE
		);

		CREATE TABLE IF NOT EXISTS signed_certs (
			id SERIAL PRIMARY KEY,
			node_name TEXT NOT NULL UNIQUE,
			csr_signature TEXT UNIQUE,
			signed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			data TEXT NOT NULL UNIQUE
		);
		`
	}

	return ""
}

// Constructor for database engine
func NewDatabaseEngine(databaseConfig *config.DatabaseConfig) (*DatabaseEngine, error) {
	var engine DatabaseEngine

	db, err := sql.Open(databaseConfig.Type, databaseConfig.ToDSN())
	if err != nil {
		return &engine, err
	}

	_, err = db.Exec(databaseSchema(databaseConfig.Type))
	if err != nil {
		return &engine, err
	}

	engine.DB = db
	engine.DBType = databaseConfig.Type

	return &engine, nil
}
