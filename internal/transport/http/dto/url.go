// Package dto предоставляет структуры данных для передачи между слоями приложения.
//
// Содержит запросы и ответы для API-эндпоинтов.
package dto

// CreateURLRequest представляет запрос на создание короткой URL.
type CreateURLRequest struct {
	// URL — исходный URL для сокращения.
	URL string `json:"url" binding:"required,url"`
}

// CreateURLResponse представляет ответ на создание короткой URL.
type CreateURLResponse struct {
	// Result — сгенерированная короткая ссылка.
	Result string `json:"result"`
}

// URLRow представляет строку URL в хранилище.
type URLRow struct {
	// ShortURL — короткий идентификатор.
	ShortURL string `json:"short_url"`
	// OriginalURL — исходный URL.
	OriginalURL string `json:"original_url"`
	// UserID — идентификатор пользователя (опционально).
	UserID string `json:"user_id,omitempty"`
}

// PingResponse представляет ответ на проверку состояния сервиса.
type PingResponse struct {
	// Status — статус ("ok" или "error").
	Status string `json:"status"`
	// Message — сообщение о состоянии.
	Message string `json:"message"`
	// Error — описание ошибки (если статус error).
	Error string `json:"error,omitempty"`
}

// CreateBatchURLRequestItem представляет один элемент пакетного запроса.
type CreateBatchURLRequestItem struct {
	// CorrelationID — идентификатор для сопоставления запроса и ответа.
	CorrelationID string `json:"correlation_id"`
	// OriginalURL — исходный URL для сокращения.
	OriginalURL string `json:"original_url"`
}

// CreateBatchURLRequest представляет пакетный запрос на создание нескольких URL.
type CreateBatchURLRequest []CreateBatchURLRequestItem

// BatchURLResponseItem представляет один элемент пакетного ответа.
type BatchURLResponseItem struct {
	// CorrelationID — идентификатор для сопоставления запроса и ответа.
	CorrelationID string `json:"correlation_id"`
	// ShortURL — сгенерированная короткая ссылка.
	ShortURL string `json:"short_url"`
}

// CreateBatchURLResponse представляет пакетный ответ с короткими ссылками.
type CreateBatchURLResponse []BatchURLResponseItem

// URLResponse представляет URL в ответе API.
type URLResponse struct {
	// ShortURL — короткая ссылка.
	ShortURL string `json:"short_url"`
	// OriginalURL — исходный URL.
	OriginalURL string `json:"original_url"`
}

// UserURLsResponse представляет список URL пользователя.
type UserURLsResponse []URLResponse

// DeleteURLsRequest представляет запрос на удаление URL.
type DeleteURLsRequest []string
