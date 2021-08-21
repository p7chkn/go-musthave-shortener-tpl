package shortener

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"
)

func AddURL(longURL string, data url.Values) string {
	shortURL := shorterURL(longURL)
	data.Set(shortURL, longURL)
	return shortURL
}

func GetURL(shortURL string, data url.Values) (string, error) {
	result := data.Get(shortURL)
	if result == "" {
		return "", errors.New("not found")
	}
	return result, nil
}

func shorterURL(longURL string) string {
	splitURL := strings.Split(longURL, "://")
	hasher := sha1.New()
	hasher.Write([]byte(splitURL[1]))
	urlHash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return string(urlHash)
}
