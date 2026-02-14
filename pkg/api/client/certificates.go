package client

import (
	"errors"
	"fmt"

	"github.com/peeklapp/peekl/pkg/api/requests"
	"github.com/peeklapp/peekl/pkg/api/responses"
)

func (c *Client) GetRootCA() (string, error) {
	endpoint := "/v1/certificates/root"
	var resp responses.GetRootCA

	err := c.get(endpoint, &resp)
	if err != nil {
		if errors.Is(err, HttpError{}) {
			detailedError, _ := err.(HttpError)
			return "", fmt.Errorf("Status code : %d. Details : %+v", detailedError.StatusCode, detailedError.ErrorBody)
		} else {
			return "", err
		}
	}

	return resp.Certificate, nil
}

func (c *Client) SubmitCertificateRequest(nodeName string, csr string) error {
	endpoint := "/v1/certificates/submit"
	body := requests.SubmitCertificateRequest{NodeName: nodeName, CSR: csr}
	var resp responses.MessageResponse

	err := c.post(endpoint, body, &resp)
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

func (c *Client) RetrieveSignedCertificate(nodeName string, csr string) (string, error) {
	endpoint := "/v1/certificates/retrieve"
	body := requests.RetrieveSignedCertificate{NodeName: nodeName, CSR: csr}
	var resp responses.RetrieveSignedCertificate

	err := c.post(endpoint, body, &resp)
	if err != nil {
		if errors.Is(err, HttpError{}) {
			detailedError, _ := err.(HttpError)
			return "", fmt.Errorf("Status code : %d. Details : %+v", detailedError.StatusCode, detailedError.ErrorBody)
		} else {
			return "", err
		}
	}

	return resp.Certificate, nil
}
