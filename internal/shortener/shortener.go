package shortener

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"
)

var data = url.Values{}

func AddURL(longURL string) string {
	shortURL := shorterURL(longURL)
	data.Set(shortURL, longURL)
	return shortURL
}

func GetURL(shortURL string) (string, error) {
	result := data.Get(shortURL)
	if result == "" {
		return "", errors.New("Not found")
	}
	return result, nil
}

func shorterURL(longURL string) string {
	splitURL := strings.Split(longURL, "://")
	hasher := sha1.New()
	hasher.Write([]byte(splitURL[1]))
	url_hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return string(url_hash)
}
