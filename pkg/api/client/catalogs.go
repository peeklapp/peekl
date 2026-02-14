package client

import (
	"errors"
	"fmt"

	"github.com/peeklapp/peekl/pkg/api/responses"
	"github.com/peeklapp/peekl/pkg/models"
)

func (c *Client) GetCatalog() ([]models.Resource, []models.Role, []string, map[string]any, error) {
	endpoint := "/v1/catalogs/catalog"
	var resp responses.GetCatalog

	err := c.get(endpoint, &resp)
	if err != nil {
		if errors.As(err, &HttpError{}) {
			detailedError, _ := err.(HttpError)
			return resp.GlobalResource, resp.Roles, resp.Tags, resp.Variables, fmt.Errorf("Status code : %d. Details : %+v", detailedError.StatusCode, detailedError.ErrorBody)
		} else {
			return resp.GlobalResource, resp.Roles, resp.Tags, resp.Variables, err
		}
	}

	return resp.GlobalResource, resp.Roles, resp.Tags, resp.Variables, nil
}
