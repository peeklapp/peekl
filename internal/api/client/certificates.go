package client

import (
	"errors"
	"fmt"

	"github.com/peeklapp/peekl/internal/api/requests"
	"github.com/peeklapp/peekl/internal/api/responses"
)

func (c *Client) EnrollAgent(csr string, token string) (string, string, error) {
	endpoint := "/v1/certificates/enroll"
	body := requests.EnrollAgent{CSR: csr, Token: token}
	var resp responses.EnrollAgent

	err := c.post(endpoint, body, &resp)
	if err != nil {
		if errors.As(err, &HttpError{}) {
			detailedError, _ := err.(HttpError)
			return "", "", fmt.Errorf("status code : %d. details : %+v", detailedError.StatusCode, detailedError.ErrorBody)
		} else {
			return "", "", err
		}
	}

	return resp.Certificate, resp.CA, nil
}
