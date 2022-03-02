package shortener

import (
	"crypto/sha1"
	"encoding/base64"
	"strings"
)

func ShorterURL(longURL string) string {
	splitURL := strings.Split(longURL, "://")
	hasher := sha1.New()
	if len(splitURL) < 2 {
		hasher.Write([]byte(longURL))
	} else {
		hasher.Write([]byte(splitURL[1]))
	}
	urlHash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return urlHash
}
