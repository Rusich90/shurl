package validators

import (
	"fmt"
	"net/url"
	"strings"
)

func ValidateCreateURLRequest(urlField string) error {
	urlField = strings.TrimSpace(urlField)
	if urlField == "" {
		return fmt.Errorf("url is required")
	}

	parsedURL, err := url.Parse(urlField)
	if err != nil {
		return fmt.Errorf("invalid url format")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("url must use http or https scheme")
	}

	return nil
}

func ValidateDeleteURLsRequest(ids []string) error {
	if len(ids) == 0 {
		return fmt.Errorf("at least one id is required")
	}

	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			return fmt.Errorf("id cannot be empty")
		}
	}

	return nil
}

func ValidateCreateBatchURLRequest(correlationID, urlField string) error {
	correlationID = strings.TrimSpace(correlationID)
	if correlationID == "" {
		return fmt.Errorf("correlation_id is required")
	}

	return ValidateCreateURLRequest(urlField)
}
