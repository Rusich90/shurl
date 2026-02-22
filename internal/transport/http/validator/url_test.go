package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCreateURLRequest(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected error
	}{
		{
			name:     "valid http url",
			url:      "http://example.com",
			expected: nil,
		},
		{
			name:     "valid https url",
			url:      "https://example.com",
			expected: nil,
		},
		{
			name:     "valid url with path",
			url:      "https://example.com/path/to/resource",
			expected: nil,
		},
		{
			name:     "valid url with query params",
			url:      "https://example.com?param1=value1&param2=value2",
			expected: nil,
		},
		{
			name:     "valid url with fragment",
			url:      "https://example.com#section",
			expected: nil,
		},
		{
			name:     "valid url with port",
			url:      "https://example.com:8080/path",
			expected: nil,
		},
		{
			name:     "valid url with auth",
			url:      "https://user:pass@example.com",
			expected: nil,
		},
		{
			name:     "empty url",
			url:      "",
			expected: assert.AnError,
		},
		{
			name:     "whitespace only",
			url:      "   ",
			expected: assert.AnError,
		},
		{
			name:     "invalid url format",
			url:      "not a url",
			expected: assert.AnError,
		},
		{
			name:     "ftp scheme not allowed",
			url:      "ftp://example.com",
			expected: assert.AnError,
		},
		{
			name:     "mailto scheme not allowed",
			url:      "mailto:test@example.com",
			expected: assert.AnError,
		},
		{
			name:     "file scheme not allowed",
			url:      "file:///path/to/file",
			expected: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreateURLRequest(tt.url)
			if tt.expected == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestValidateDeleteURLsRequest(t *testing.T) {
	tests := []struct {
		name     string
		ids      []string
		expected error
	}{
		{
			name:     "valid single id",
			ids:      []string{"abc123"},
			expected: nil,
		},
		{
			name:     "valid multiple ids",
			ids:      []string{"abc123", "def456", "ghi789"},
			expected: nil,
		},
		{
			name:     "valid ids with whitespace",
			ids:      []string{" abc123 ", "def456"},
			expected: nil,
		},
		{
			name:     "empty slice",
			ids:      []string{},
			expected: assert.AnError,
		},
		{
			name:     "id with empty string",
			ids:      []string{"abc123", ""},
			expected: assert.AnError,
		},
		{
			name:     "id with whitespace only",
			ids:      []string{"abc123", "   "},
			expected: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDeleteURLsRequest(tt.ids)
			if tt.expected == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestValidateCreateBatchURLRequest(t *testing.T) {
	tests := []struct {
		name     string
		corID    string
		url      string
		expected error
	}{
		{
			name:     "valid request",
			corID:    "123",
			url:      "https://example.com",
			expected: nil,
		},
		{
			name:     "valid request with whitespace",
			corID:    " 123 ",
			url:      " https://example.com ",
			expected: nil,
		},
		{
			name:     "empty correlation id",
			corID:    "",
			url:      "https://example.com",
			expected: assert.AnError,
		},
		{
			name:     "whitespace correlation id",
			corID:    "   ",
			url:      "https://example.com",
			expected: assert.AnError,
		},
		{
			name:     "empty url",
			corID:    "123",
			url:      "",
			expected: assert.AnError,
		},
		{
			name:     "invalid url",
			corID:    "123",
			url:      "not a url",
			expected: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreateBatchURLRequest(tt.corID, tt.url)
			if tt.expected == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
