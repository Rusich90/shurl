package model

type CreateURLRequest struct {
	URL string `json:"url" binding:"required,url"`
}

type CreateURLResponse struct {
	Result string `json:"result"`
}

type URLRow struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
