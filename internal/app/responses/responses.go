package responses

type PostURL struct {
	URL string
}

type ManyPostURL struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ManyPostResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type GetURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type StatResponse struct {
	CountURL  int `json:"urls"`
	CountUser int `json:"users"`
}
