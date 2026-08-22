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
	"os"
	"time"

	"github.com/peeklapp/peekl/internal/utils"
)

func CreateCertificate(dnsNames []string, caFilePath string, caKeyPath string, outCertFilePath string, outKeyFilePath string) error {
	// Make sure peekl is in the DNS names
	dnsNames = append(dnsNames, "peekl")

	// Create cert values
	certValues := x509.Certificate{
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 365 * 10),
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
	loadedCa, err := LoadCertificateFromFile(caFilePath)
	if err != nil {
		return err
	}

	// Load CA key from file
	loadedCaKey, err := LoadPKCS8PrivateKeyFromFile(caKeyPath)
	if err != nil {
		return err
	}

	// Get public key
	certPubKey := certPrivKey.Public()

	// Generate actual certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, &certValues, loadedCa, certPubKey, loadedCaKey)
	if err != nil {
		return err
	}

	// Create CRT file on disk
	crtOut, err := os.OpenFile(outCertFilePath, os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer utils.CloseWithoutError(crtOut)

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
	csrKeyOut, err := os.OpenFile(outKeyFilePath, os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer utils.CloseWithoutError(csrKeyOut)

	// Write private key to file
	if err := pem.Encode(csrKeyOut, &pem.Block{Type: "ECDSA PRIVATE KEY", Bytes: marshalledCsrKey}); err != nil {
		return err
	}

	return nil
}

func CreateCertificateSigningRequest(nodeName string, keyFileOutput string, csrFileOutput string) error {
	// Generate private key
	curve := elliptic.P384()
	csrKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return err
	}

	// Marshall private key in writable data
	marshalledCsrKey, err := x509.MarshalPKCS8PrivateKey(csrKey)
	if err != nil {
		return err
	}

	// Create private key file on disk
	csrKeyOut, err := os.OpenFile(keyFileOutput, os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	// Write private key to file
	if err := pem.Encode(csrKeyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: marshalledCsrKey}); err != nil {
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
	csrOut, err := os.OpenFile(csrFileOutput, os.O_CREATE, 0600)
	if err != nil {
		return nil
	}
	defer utils.CloseWithoutError(csrOut)

	// Write CSR to file
	if err := pem.Encode(csrOut, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes}); err != nil {
		return err
	}

	return nil
}

func LoadCertificateFromData(data string) (*x509.Certificate, error) {
	certificatePem, _ := pem.Decode([]byte(data))
	if certificatePem == nil {
		return &x509.Certificate{}, fmt.Errorf("could not decode certificate bytes")
	}

	certificate, err := x509.ParseCertificate(certificatePem.Bytes)
	if err != nil {
		return &x509.Certificate{}, err
	}

	return certificate, nil
}

func LoadCertificateFromFile(certificateFile string) (*x509.Certificate, error) {
	certificateBytes, err := os.ReadFile(certificateFile)
	if err != nil {
		return &x509.Certificate{}, err
	}

	certificatePem, _ := pem.Decode(certificateBytes)
	if certificatePem == nil {
		return &x509.Certificate{}, fmt.Errorf("could not decode certificate bytes")
	}

	certificate, err := x509.ParseCertificate(certificatePem.Bytes)
	if err != nil {
		return &x509.Certificate{}, err
	}

	return certificate, nil
}

func LoadPKCS8PrivateKeyFromFile(privateKeyFile string) (*ecdsa.PrivateKey, error) {
	privateKeyBytes, err := os.ReadFile(privateKeyFile)
	if err != nil {
		return &ecdsa.PrivateKey{}, err
	}

	privPem, _ := pem.Decode(privateKeyBytes)
	if privPem == nil {
		return &ecdsa.PrivateKey{}, fmt.Errorf("could not decode EC private key bytes")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(privPem.Bytes)
	if err != nil {
		return &ecdsa.PrivateKey{}, err
	}

	switch privateKey := privateKey.(type) {
	case *ecdsa.PrivateKey:
		return privateKey, nil
	}

	return &ecdsa.PrivateKey{}, fmt.Errorf("the key was not of type ECDSA")
}

func LoadCertificateSigningRequest(csrData string) (*x509.CertificateRequest, error) {
	csrPem, _ := pem.Decode([]byte(csrData))
	if csrPem == nil {
		return &x509.CertificateRequest{}, fmt.Errorf("could not decode CSR Bytes")
	}

	certificateRequest, err := x509.ParseCertificateRequest(csrPem.Bytes)
	if err != nil {
		return &x509.CertificateRequest{}, err
	}

	return certificateRequest, nil
}

func SignCertificateSigningRequest(csrData string, caFilePath string, caKeyPath string) (string, error) {
	loadedCsr, err := LoadCertificateSigningRequest(csrData)
	if err != nil {
		return "", err
	}

	loadedCa, err := LoadCertificateFromFile(caFilePath)
	if err != nil {
		return "", err
	}
	loadedCaKey, err := LoadPKCS8PrivateKeyFromFile(caKeyPath)
	if err != nil {
		return "", err
	}

	certTemplate := x509.Certificate{
		Subject:               loadedCsr.Subject,
		DNSNames:              loadedCsr.DNSNames,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 365),
		IsCA:                  false,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &certTemplate, loadedCa, loadedCsr.PublicKey, loadedCaKey)
	if err != nil {
		return "", err
	}

	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})), nil
}

func GetCertificateSigningRequestSignature(csr string) string {
	hash := sha256.Sum256([]byte(csr))
	return hex.EncodeToString(hash[:])
}
