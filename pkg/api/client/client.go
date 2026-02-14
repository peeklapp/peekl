package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/peeklapp/peekl/pkg/api/responses"
	"github.com/peeklapp/peekl/pkg/config"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewApiClient(conf config.AgentConfig, bootstrap bool, certPool *x509.CertPool) (*Client, error) {
	var httpClient http.Client
	if bootstrap {
		// Create unsecure HTTP client, only used for bootstrap
		httpClient = http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
					RootCAs:            certPool,
				},
			},
		}
	} else {
		// Load the CA certificate
		caCertPool := x509.NewCertPool()
		caCert, err := os.ReadFile(conf.Certificates.CaFilePath)
		if err != nil {
			return &Client{}, err
		}
		caCertPool.AppendCertsFromPEM(caCert)

		// Load the certificate and the key of the agent
		cert, err := tls.LoadX509KeyPair(conf.Certificates.CertificateFilePath, conf.Certificates.CertificateKeyPath)
		if err != nil {
			return &Client{}, err
		}

		// Create HTTP client with certificate and CA for proper mTLS
		httpClient = http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:      caCertPool,
					Certificates: []tls.Certificate{cert},
				},
			},
		}
	}

	// Create URL on which to contact server
	baseUrl := fmt.Sprintf("https://%s:%d", conf.Server.Host, conf.Server.Port)

	// Return actual API client
	return &Client{
		baseURL:    baseUrl,
		httpClient: &httpClient,
	}, nil
}

// Represent an advanced HTTP error with body and status code
type HttpError struct {
	StatusCode int
	ErrorBody  responses.ErrorResponse
}

func (h HttpError) Error() string {
	return fmt.Sprintf("Response is not OK : %d", h.StatusCode)
}

// Make a get request
func (c *Client) get(endpoint string, out any) error {
	// Compute URL
	url := c.baseURL + endpoint

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// Process request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Make sure response is ok, if not process it
	if resp.StatusCode > 299 {
		// Get body of request (contains error details)
		var errorResponse responses.ErrorResponse
		err := json.NewDecoder(resp.Body).Decode(&errorResponse)
		if err != nil {
			return err
		}

		// Return proper response
		httpError := HttpError{ErrorBody: errorResponse, StatusCode: resp.StatusCode}
		return httpError
	}

	// Write result to passed variable
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}

	return nil
}

// Make a post request
func (c *Client) post(endpoint string, body any, out any) error {
	url := c.baseURL + endpoint

	// Serialize body to JSON
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		// Get body of request (contains error details)
		var errorResponse responses.ErrorResponse
		err := json.NewDecoder(resp.Body).Decode(&errorResponse)
		if err != nil {
			return err
		}

		// Return proper response
		httpError := HttpError{ErrorBody: errorResponse, StatusCode: resp.StatusCode}
		return httpError
	}

	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}

	return nil
}
