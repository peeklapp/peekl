package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/peeklapp/peekl/internal/config"
	"github.com/peeklapp/peekl/internal/models"
	_ "modernc.org/sqlite"
)

type DatabaseEngine struct {
	db     *sql.DB
	dbType string
}

// Signed certificates
func (c *DatabaseEngine) InsertSignedCertificate(nodeName string, certificate string) error {
	queries := map[string]string{
		"sqlite":   "INSERT INTO signed_certs (node_name, certificate) VALUES (?, ?);",
		"postgres": "INSERT INTO signed_certs (node_name, certificate) VALUES ($1, $2);",
	}
	_, err := c.db.Exec(queries[c.dbType], nodeName, certificate)
	if err != nil {
		return fmt.Errorf("could not insert signed certificate in database : %s", err.Error())
	}
	return nil
}

func (c *DatabaseEngine) GetSignedCertificateByNodeName(nodeName string) (models.SignedCertificate, error) {
	queries := map[string]string{
		"sqlite":   "SELECT node_name, signed_at, certificate FROM signed_certs WHERE node_name = ?;",
		"postgres": "SELECT node_name, signed_at, certificate FROM signed_certs WHERE node_name = $1;",
	}
	var signedCert models.SignedCertificate
	err := c.db.QueryRow(
		queries[c.dbType], nodeName,
	).Scan(&signedCert.NodeName, &signedCert.SignedAt, &signedCert.Certificate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return signedCert, models.SignedCertificateNotFoundByNodeName{NodeName: nodeName}
		}
		return signedCert, err
	}
	return signedCert, nil
}

func (c *DatabaseEngine) ListSignedCertificates() ([]models.SignedCertificate, error) {
	var signedCerts []models.SignedCertificate

	rows, err := c.db.Query("SELECT node_name, signed_at FROM signed_certs")
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

// Revoked certificates
func (c *DatabaseEngine) DeleteSignedCertificate(nodeName string) error {
	queries := map[string]string{
		"sqlite":   "DELETE FROM signed_certs WHERE node_name = ?;",
		"postgres": "DELETE FROM signed_certs WHERE node_name = $1;",
	}
	_, err := c.db.Exec(queries[c.dbType], nodeName)
	if err != nil {
		return err
	}
	return nil
}

func (c *DatabaseEngine) InsertRevokedCertificate(nodeName string, serialNumber string) error {
	queries := map[string]string{
		"sqlite":   "INSERT INTO revoked_certs (node_name, serial_number) VALUES (?, ?);",
		"postgres": "INSERT INTO revoked_certs (node_name, serial_number) VALUES ($1, $2);",
	}
	_, err := c.db.Exec(queries[c.dbType], nodeName, serialNumber)
	if err != nil {
		return fmt.Errorf("could not insert revoked certificate in database : %s", err.Error())
	}
	return nil
}

func (c *DatabaseEngine) ListRevokedCertificates() ([]models.RevokedCertificate, error) {
	var revokedCerts []models.RevokedCertificate

	rows, err := c.db.Query("SELECT node_name, serial_number, revoked_at FROM revoked_certs")
	if err != nil {
		return revokedCerts, err
	}
	defer rows.Close() //nolint:errcheck

	for rows.Next() {
		var r models.RevokedCertificate
		err := rows.Scan(&r.NodeName, &r.SerialNumber, &r.RevokedAt)
		if err != nil {
			return revokedCerts, err
		}
		revokedCerts = append(revokedCerts, r)
	}

	return revokedCerts, nil
}

func (c *DatabaseEngine) GetRevokedCertificate(serialNumber string) (models.RevokedCertificate, error) {
	queries := map[string]string{
		"sqlite":   "SELECT node_name, serial_number, revoked_at FROM revoked_certs WHERE serial_number = ?;",
		"postgres": "SELECT node_name, serial_number, revoked_at FROM revoked_certs WHERE serial_number = $1;",
	}
	var revokedCert models.RevokedCertificate
	err := c.db.QueryRow(
		queries[c.dbType], serialNumber,
	).Scan(&revokedCert.NodeName, &revokedCert.SerialNumber, &revokedCert.RevokedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return revokedCert, models.RevokedCertificateNotFound{SerialNumber: serialNumber}
		}
		return revokedCert, err
	}

	return revokedCert, nil
}

func (c *DatabaseEngine) IsNodeNameUsedInSigned(nodeName string) (bool, error) {
	queries := map[string]string{
		"sqlite":   "SELECT node_name FROM signed_certs WHERE node_name = ?;",
		"postgres": "SELECT node_name FROM signed_certs WHERE node_name = $1;",
	}
	err := c.db.QueryRow(queries[c.dbType], nodeName).Scan(&nodeName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *DatabaseEngine) IsNodeNameUsedInPending(nodeName string) (bool, error) {
	queries := map[string]string{
		"sqlite":   "SELECT node_name FROM pending_certs WHERE node_name = ?;",
		"postgres": "SELECT node_name FROM pending_certs WHERE node_name = $1;",
	}
	err := c.db.QueryRow(queries[c.dbType], nodeName).Scan(&nodeName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *DatabaseEngine) InsertEnrollmentToken(tokenHash string, ip string, valid_until time.Time) error {
	queries := map[string]string{
		"sqlite":   "INSERT INTO enrollment_tokens (ip, token_hash, valid_until) VALUES (?, ?, ?);",
		"postgres": "INSERT INTO enrollment_tokens (ip, token_hash, valid_until) VALUES ($1, $2, $3);",
	}
	_, err := c.db.Exec(queries[c.dbType], ip, tokenHash, valid_until)
	if err != nil {
		return fmt.Errorf("could not insert enrollment token in database : %s", err.Error())
	}
	return nil
}

func (c *DatabaseEngine) GetEnrollmentToken(ip string) (models.EnrollmentToken, error) {
	queries := map[string]string{
		"sqlite":   "SELECT ip, token_hash, created_at, valid_until FROM enrollment_tokens WHERE ip = ?;",
		"postgres": "SELECT ip, token_hash, created_at, valid_until FROM enrollment_tokens WHERE ip = $1;",
	}
	var enrollmentToken models.EnrollmentToken
	err := c.db.QueryRow(
		queries[c.dbType], ip,
	).Scan(&enrollmentToken.Ip, &enrollmentToken.TokenHash, &enrollmentToken.CreatedAt, &enrollmentToken.ValidUntil)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return enrollmentToken, models.EnrollmentTokenNotFound{}
		}
		return enrollmentToken, err
	}
	return enrollmentToken, nil
}

func (c *DatabaseEngine) DeleteEnrollmentToken(ip string) error {
	queries := map[string]string{
		"sqlite":   "DELETE FROM enrollment_tokens WHERE ip = ?;",
		"postgres": "DELETE FROM enrollment_tokens WHERE ip = $1;",
	}
	_, err := c.db.Exec(queries[c.dbType], ip)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.EnrollmentTokenNotFound{}
		}
		return err
	}
	return nil
}

func (c *DatabaseEngine) ListEnrollmentToken() ([]models.EnrollmentToken, error) {
	var enrollmentTokens []models.EnrollmentToken

	rows, err := c.db.Query("SELECT ip, token_hash, created_at, valid_until FROM enrollment_tokens WHERE valid_until < CURRENT_TIMESTAMP;")
	if err != nil {
		return enrollmentTokens, err
	}
	defer rows.Close() //nolint:errcheck

	for rows.Next() {
		var t models.EnrollmentToken
		err := rows.Scan(&t.Ip, &t.TokenHash, &t.CreatedAt, &t.ValidUntil)
		if err != nil {
			return enrollmentTokens, err
		}
		enrollmentTokens = append(enrollmentTokens, t)
	}
	return enrollmentTokens, nil
}

func (c *DatabaseEngine) DoesATokenExistAndIsValid(ip string) (bool, error) {
	queries := map[string]string{
		"sqlite":   "SELECT COUNT(*) FROM enrollment_tokens WHERE ip = ? AND valid_until < ?;",
		"postgres": "SELECT COUNT(*) FROM enrollment_tokens WHERE ip = $1 AND valid_until < $2;",
	}
	currentTime := time.Now()

	var tokenFound int
	err := c.db.QueryRow(queries[c.dbType], ip, currentTime).Scan(&tokenFound)
	if err != nil {
		return false, err
	}

	if tokenFound > 0 {
		return true, nil
	}
	return false, nil
}

func databaseSchema(dbType string) string {
	switch dbType {
	case "sqlite":
		return `
    CREATE TABLE IF NOT EXISTS signed_certs (
    	id INTEGER PRIMARY KEY AUTOINCREMENT,
    	node_name TEXT NOT NULL UNIQUE,
    	signed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    	certificate TEXT NOT NULL UNIQUE
    );

    CREATE TABLE IF NOT EXISTS revoked_certs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			node_name TEXT NOT NULL,
			serial_number TEXT NOT NULL UNIQUE,
			revoked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

		CREATE TABLE IF NOT EXISTS enrollment_tokens (
			ip TEXT NOT NULL PRIMARY KEY,
			token_hash TEXT NOT NULL,
			valid_until TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		`
	case "postgres":
		return `
		CREATE TABLE IF NOT EXISTS signed_certs (
			id SERIAL PRIMARY KEY,
			node_name TEXT NOT NULL UNIQUE,
			signed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			certificate TEXT NOT NULL UNIQUE
		);

		CREATE TABLE IF NOT EXISTS revoked_certs (
			id SERIAL PRIMARY KEY,
			node_name TEXT NOT NULL,
			serial_number TEXT NOT NULL UNIQUE,
			revoked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS enrollment_tokens (
			ip TEXT NOT NULL PRIMARY KEY,
			token_hash TEXT NOT NULL,
			valid_until TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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

	engine.db = db
	engine.dbType = databaseConfig.Type

	return &engine, nil
}
