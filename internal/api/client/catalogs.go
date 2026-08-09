package client

import (
	"errors"
	"fmt"

	"github.com/peeklapp/peekl/internal/api/requests"
	"github.com/peeklapp/peekl/internal/api/responses"
)

func (c *Client) InquiryForCatalog(environment string) (string, string, string, string, error) {
	endpoint := "/v1/catalogs/catalog"
	body := requests.InquiryForCatalog{Environment: environment}
	var resp responses.InquiryForCatalog

	err := c.post(endpoint, body, &resp)
	if err != nil {
		if errors.As(err, &HttpError{}) {
			detailedError, _ := err.(HttpError)
			return "", "", "", "", fmt.Errorf("Status code : %d. Details : %+v", detailedError.StatusCode, detailedError.ErrorBody)
		} else {
			return "", "", "", "", err
		}
	}

	return resp.NodeTarball.Path, resp.NodeTarball.Hash, resp.CodeTarball.Path, resp.CodeTarball.Hash, nil
}
