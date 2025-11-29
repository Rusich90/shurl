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
