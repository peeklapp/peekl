package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redat00/peekl/pkg/api/responses"
	"github.com/redat00/peekl/pkg/config"
)

type Client struct {
	baseURL            string
	httpClient         *http.Client
	insecureHttpClient *http.Client
}

func NewApiClient(conf config.AgentServerConfig) *Client {
	baseUrl := fmt.Sprintf("https://%s:%d", conf.Host, conf.Port)

	httpClient := http.Client{
		Timeout: 10 * time.Second,
	}

	insecureTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	insecureHttpClient := http.Client{
		Timeout:   10 * time.Second,
		Transport: insecureTransport,
	}

	return &Client{
		baseURL:            baseUrl,
		httpClient:         &httpClient,
		insecureHttpClient: &insecureHttpClient,
	}
}

// Represent an advanced HTTP error with body and status code
type HttpError struct {
	StatusCode int
	ErrorBody  responses.ErrorResponse
}

func (h HttpError) Error() string {
	return fmt.Sprintf("Reponse is not OK : %d", h.StatusCode)
}

// Make a get request
func (c *Client) get(endpoint string, result any, unsecure bool) error {
	// Compute URL
	url := c.baseURL + endpoint

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// Process request
	var resp *http.Response
	if !unsecure {
		resp, err = c.httpClient.Do(req)
		if err != nil {
			return err
		}
	} else {
		resp, err = c.insecureHttpClient.Do(req)
		if err != nil {
			return err
		}
	}
	defer resp.Body.Close()

	// Make sure response is ok, if not process it
	if resp.StatusCode != http.StatusOK {
		// Get body of request (contains error details)
		var errorResponse responses.ErrorResponse
		err := json.NewDecoder(resp.Body).Decode(&errorResponse)
		if err != nil {
			return err
		}

		// Return proper response
		httpError := HttpError{ErrorBody: errorResponse}
		return httpError
	}

	// Write result to passed variable
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return err
	}

	return nil
}

// Make a post request
func (c *Client) post(endpoint string, body any, result any) error {
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

	if resp.StatusCode != http.StatusOK {
		// Get body of request (contains error details)
		var errorResponse responses.ErrorResponse
		err := json.NewDecoder(resp.Body).Decode(&errorResponse)
		if err != nil {
			return err
		}

		// Return proper response
		httpError := HttpError{ErrorBody: errorResponse}
		return httpError
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return err
		}
	}

	return nil
}
