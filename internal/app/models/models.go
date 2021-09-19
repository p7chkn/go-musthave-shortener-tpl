package models

type ResponseGetURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

//go:generate mockery --name=RepositoryInterface
type RepositoryInterface interface {
	AddURL(longURL string, shortURL string, user string) error
	GetURL(shortURL string) (string, error)
	GetUserURL(user string) []ResponseGetURL
	Ping() error
}
