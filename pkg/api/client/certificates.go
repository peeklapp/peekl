package client

import (
	"errors"
	"fmt"

	"github.com/redat00/peekl/pkg/api/requests"
	"github.com/redat00/peekl/pkg/api/responses"
)

func (c *Client) SubmitCertificateRequest(nodeName string, csr string) error {
	endpoint := "/v1/certificates/submit"
	body := requests.SubmitCertificateRequest{NodeName: nodeName, CSR: csr}
	var resp responses.MessageResponse

	err := c.post(endpoint, body, resp)
	if err != nil {
		if errors.Is(err, HttpError{}) {
			detailedError, _ := err.(HttpError)
			return fmt.Errorf("Status code : %d. Details : %+v", detailedError.StatusCode, detailedError.ErrorBody)
		} else {
			return err
		}
	}

	return nil
}

func (c *Client) RetrieveSignedCertificate(nodeName string, csr string) (*responses.RetrieveSignedCertificate, error) {
	endpoint := "/v1/certificates/retrieve"
	body := requests.RetrieveSignedCertificate{NodeName: nodeName, CSR: csr}
	var resp responses.RetrieveSignedCertificate

	err := c.post(endpoint, body, resp)
	if err != nil {
		if errors.Is(err, HttpError{}) {
			detailedError, _ := err.(HttpError)
			return &resp, fmt.Errorf("Status code : %d. Details : %+v", detailedError.StatusCode, detailedError.ErrorBody)
		} else {
			return &resp, err
		}
	}

	return &resp, nil
}
