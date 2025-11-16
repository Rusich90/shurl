package model

type CreateURLRequest struct {
	URL string `json:"url" binding:"required,url"`
}

type CreateURLResponse struct {
	Result string `json:"result"`
}
