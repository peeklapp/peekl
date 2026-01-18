package certs

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

const SERIAL_NUMBER_LIMIT = 128

func CreateCertificate(caFolder string, certFolder string, nodeName string, dnsNames []string) error {
	// Create cert values
	certValues := x509.Certificate{
		SerialNumber: big.NewInt(2019),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour * 24 * 365 * 10),
		Subject: pkix.Name{
			Organization:  []string{"Peekl"},
			Country:       []string{"FR"},
			Province:      []string{""},
			Locality:      []string{"Paris"},
			StreetAddress: []string{"Le Marais"},
			PostalCode:    []string{"75004"},
			CommonName:    "peekl",
		},
		DNSNames: dnsNames,
	}

	// Generate private key
	curve := elliptic.P384()
	certPrivKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return err
	}

	// Load CA from file
	loadedCa, err := LoadCertificateFromFile(fmt.Sprintf("%s/ca.pem", caFolder))
	if err != nil {
		return err
	}

	// Load CA key from file
	loadedCaKey, err := LoadECPrivateKeyFromFile(fmt.Sprintf("%s/ca.key", caFolder))
	if err != nil {
		return err
	}

	// Get public key
	certPubKey := certPrivKey.Public()

	// Generate actual certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, &certValues, loadedCa, certPubKey, loadedCaKey)

	// Create CRT file on disk
	crtOut, err := os.Create(filepath.Join(certFolder, fmt.Sprintf("%s.crt", nodeName)))
	if err != nil {
		return nil
	}
	defer crtOut.Close()

	// Write CRT to file
	if err := pem.Encode(crtOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		return err
	}

	// Marshall private key in writable data
	marshalledCsrKey, err := x509.MarshalECPrivateKey(certPrivKey)
	if err != nil {
		return err
	}

	// Create private key file on disk
	csrKeyOut, err := os.Create(filepath.Join(certFolder, fmt.Sprintf("%s.key", nodeName)))
	if err != nil {
		return err
	}
	defer csrKeyOut.Close()

	// Write private key to file
	if err := pem.Encode(csrKeyOut, &pem.Block{Type: "ECDSA PRIVATE KEY", Bytes: marshalledCsrKey}); err != nil {
		return err
	}

	return nil
}

func CreateCertificateSigningRequest(certFolder string, nodeName string) error {
	// Generate private key
	curve := elliptic.P384()
	csrKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return err
	}

	// Marshall private key in writable data
	marshalledCsrKey, err := x509.MarshalECPrivateKey(csrKey)
	if err != nil {
		return err
	}

	// Create private key file on disk
	csrKeyOut, err := os.Create(filepath.Join(certFolder, fmt.Sprintf("%s.key", nodeName)))
	if err != nil {
		return err
	}

	// Write private key to file
	if err := pem.Encode(csrKeyOut, &pem.Block{Type: "ECDSA PRIVATE KEY", Bytes: marshalledCsrKey}); err != nil {
		return err
	}

	// Generate CSR Data
	var csrTemplate = x509.CertificateRequest{
		Subject: pkix.Name{
			Organization:  []string{"Peekl"},
			Country:       []string{"FR"},
			Province:      []string{""},
			Locality:      []string{"Paris"},
			StreetAddress: []string{"Le Marais"},
			PostalCode:    []string{"75004"},
			CommonName:    nodeName,
		},
		DNSNames: []string{nodeName},
	}

	// Generate actual CSR
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, csrKey)
	if err != nil {
		return err
	}

	// Create CSR file on disk
	csrOut, err := os.Create(filepath.Join(certFolder, fmt.Sprintf("%s.csr", nodeName)))
	if err != nil {
		return nil
	}
	defer csrOut.Close()

	// Write CSR to file
	if err := pem.Encode(csrOut, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes}); err != nil {
		return err
	}

	return nil
}

func LoadCertificateFromFile(certificateFile string) (*x509.Certificate, error) {
	certificateBytes, err := os.ReadFile(certificateFile)
	if err != nil {
		return &x509.Certificate{}, err
	}

	certificatePem, _ := pem.Decode(certificateBytes)
	if certificatePem == nil {
		return &x509.Certificate{}, fmt.Errorf("Could not decode certificate bytes")
	}

	certificate, err := x509.ParseCertificate(certificatePem.Bytes)
	if err != nil {
		return &x509.Certificate{}, err
	}

	return certificate, nil
}

func LoadECPrivateKeyFromFile(privateKeyFile string) (*ecdsa.PrivateKey, error) {
	privateKeyBytes, err := os.ReadFile(privateKeyFile)
	if err != nil {
		return &ecdsa.PrivateKey{}, err
	}

	privPem, _ := pem.Decode(privateKeyBytes)
	if privPem == nil {
		return &ecdsa.PrivateKey{}, fmt.Errorf("Could not decode EC private key bytes")
	}

	privateKey, err := x509.ParseECPrivateKey(privPem.Bytes)
	if err != nil {
		fmt.Println("this one is failing")
		return &ecdsa.PrivateKey{}, err
	}

	return privateKey, nil
}

func LoadCertificateSigningRequestFromFile(csrFolders string, nodeName string) (*x509.CertificateRequest, error) {
	csrBytes, err := os.ReadFile(filepath.Join(csrFolders, fmt.Sprintf("%s.csr", nodeName)))
	if err != nil {
		return &x509.CertificateRequest{}, err
	}

	csrPem, _ := pem.Decode(csrBytes)
	if csrPem == nil {
		return &x509.CertificateRequest{}, fmt.Errorf("Could not decode CSR Bytes")
	}

	certificateRequest, err := x509.ParseCertificateRequest(csrPem.Bytes)
	if err != nil {
		return &x509.CertificateRequest{}, err
	}

	return certificateRequest, nil
}

func SignCertificateSigningRequest(caFolder string, csrFolder string, signedFolder string, nodeName string) error {
	// Load the CA
	loadedCa, err := LoadCertificateFromFile(fmt.Sprintf("%s/ca.pem", caFolder))
	if err != nil {
		return err
	}

	// Load the CA private key
	loadedCaKey, err := LoadECPrivateKeyFromFile(fmt.Sprintf("%s/ca.key", caFolder))
	if err != nil {
		return err
	}

	// Load the CSR
	loadedCsr, err := LoadCertificateSigningRequestFromFile(csrFolder, nodeName)
	if err != nil {
		return err
	}

	certTemplate := x509.Certificate{
		Subject:               loadedCsr.Subject,
		DNSNames:              loadedCsr.DNSNames,
		SerialNumber:          big.NewInt(2019),
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 365 * 10),
		IsCA:                  false,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	// Generate actual certificate from CSR
	certBytes, err := x509.CreateCertificate(rand.Reader, &certTemplate, loadedCa, loadedCsr.PublicKey, loadedCaKey)
	if err != nil {
		return err
	}

	// Create CRT file on disk
	crtOut, err := os.Create(filepath.Join(signedFolder, fmt.Sprintf("%s.crt", nodeName)))
	if err != nil {
		return nil
	}
	defer crtOut.Close()

	// Write CRT to file
	if err := pem.Encode(crtOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		return err
	}

	return nil
}

func GetCertificateSigningRequestSignature(csr string) string {
	hash := sha256.Sum256([]byte(csr))
	return hex.EncodeToString(hash[:])
}
