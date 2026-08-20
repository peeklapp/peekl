package client

import (
	"errors"
	"fmt"
)

func (c *Client) DownloadFile(filePath string, outputFile string) error {
	err := c.get(filePath, nil, outputFile)
	if err != nil {
		if errors.As(err, &HttpError{}) {
			detailedError, _ := err.(HttpError)
			return fmt.Errorf("status code : %d. details : %+v", detailedError.StatusCode, detailedError.ErrorBody)
		} else {
			return err
		}
	}
	return nil
}
