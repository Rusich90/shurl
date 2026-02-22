// Package validators предоставляет функции для валидации входных данных.
//
// Используется для проверки корректности URL и списков ID перед обработкой.
package validators

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateCreateURLRequest валидирует URL для создания короткой ссылки.
//
// Проверяет, что URL не пустой и использует схему http или https.
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

// ValidateDeleteURLsRequest валидирует список ID для удаления.
//
// Проверяет, что список не пустой и все ID не пустые строки.
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

// ValidateCreateBatchURLRequest валидирует элемент пакетного запроса.
//
// Проверяет, что correlation_id не пустой и URL валиден.
func ValidateCreateBatchURLRequest(correlationID, urlField string) error {
	correlationID = strings.TrimSpace(correlationID)
	if correlationID == "" {
		return fmt.Errorf("correlation_id is required")
	}

	return ValidateCreateURLRequest(urlField)
}
