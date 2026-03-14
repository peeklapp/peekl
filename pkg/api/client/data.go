package client

import (
	"errors"
	"fmt"

	"github.com/peeklapp/peekl/pkg/api/requests"
	"github.com/peeklapp/peekl/pkg/api/responses"
)

func (c *Client) RetrieveFile(filename string, environment string, roleName string) (string, error) {
	endpoint := "/v1/data/file"
	body := requests.RetrieveFile{Filename: filename, Environment: environment, RoleName: roleName}
	var resp responses.RetrieveFile

	err := c.post(endpoint, &body, &resp)
	if err != nil {
		if errors.Is(err, HttpError{}) {
			detailedError, _ := err.(HttpError)
			return "", fmt.Errorf("Status code : %d. Details : %+v", detailedError.StatusCode, detailedError.ErrorBody)
		} else {
			return "", err
		}
	}

	return resp.Content, nil
}

func (c *Client) RetrieveTemplate(templateName string, environment string, roleName string) (string, error) {
	endpoint := "/v1/data/template"
	body := requests.RetrieveTemplate{TemplateName: templateName, Environment: environment, RoleName: roleName}
	var resp responses.RetrieveTemplate

	err := c.post(endpoint, &body, &resp)
	if err != nil {
		if errors.Is(err, HttpError{}) {
			detailedError, _ := err.(HttpError)
			return "", fmt.Errorf("Status code : %d. Details: %+v", detailedError.StatusCode, detailedError.ErrorBody)
		} else {
			return "", err
		}
	}

	return resp.Content, nil
}
