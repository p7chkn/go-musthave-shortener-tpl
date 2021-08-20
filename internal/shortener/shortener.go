package shortener

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"
)

var data = url.Values{}

func AddUrl(longUrl string) string {
	shortUrl := shorterURL(longUrl)
	data.Set(shortUrl, longUrl)
	return shortUrl
}

func GetUrl(shortUrl string) (string, error) {
	result := data.Get(shortUrl)
	if result == "" {
		return "", errors.New("Not found")
	}
	return result, nil
}

func shorterURL(longUrl string) string {
	splitUrl := strings.Split(longUrl, "://")
	hasher := sha1.New()
	hasher.Write([]byte(splitUrl[1]))
	url_hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return string(url_hash)
}
