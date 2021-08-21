package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/shortener"
)

func response(result map[string]string) []byte {
	if len(result) == 0 {
		return []byte{}
	}
	response, _ := json.Marshal(result)
	return response
}

func RetriveShortURL(data url.Values) func(c *gin.Context) {
	return func(c *gin.Context) {
		result := map[string]string{}
		long, err := shortener.GetURL(c.Param("id"), data)

		if err != nil {
			result["detail"] = err.Error()
			c.IndentedJSON(http.StatusNotFound, result)
			return
		}

		c.Header("Location", long)
		c.String(http.StatusTemporaryRedirect, "")
	}
}

func CreateShortURL(data url.Values) func(c *gin.Context) {
	return func(c *gin.Context) {
		result := map[string]string{}
		defer c.Request.Body.Close()

		body, err := ioutil.ReadAll(c.Request.Body)

		if err != nil {
			result["detail"] = "Bad request"
			c.IndentedJSON(http.StatusBadRequest, result)
			return
		}
		short := shortener.AddURL(string(body), data)
		c.String(http.StatusCreated, "http://localhost:8080/"+short)
	}
}

func URLHandler(data url.Values) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		result := map[string]string{}

		returnResult := func() {
			w.Write(response(result))
		}

		defer returnResult()

		w.Header().Set("Content-Type", "application/json")

		f := func(c rune) bool {
			return c == '/'
		}

		splitURL := strings.FieldsFunc(r.URL.Path, f)

		if len(splitURL) > 1 {
			w.WriteHeader(http.StatusNotFound)
			result["detail"] = "Page not found"
			return
		}

		switch r.Method {

		case http.MethodGet:

			if len(splitURL) < 1 {
				w.WriteHeader(http.StatusBadRequest)
				result["detail"] = "Bad request"
				return
			}

			long, err := shortener.GetURL(splitURL[0], data)

			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				result["detail"] = "Not found"
				return
			}
			w.Header().Set("Location", long)
			w.WriteHeader(http.StatusTemporaryRedirect)

		case http.MethodPost:

			if len(splitURL) > 0 {
				w.WriteHeader(http.StatusBadRequest)
				result["detail"] = "Bad request"
				return
			}

			defer r.Body.Close()

			body, err := ioutil.ReadAll(r.Body)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				result["detail"] = "Bad request"
				return
			}

			short := shortener.AddURL(string(body), data)
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("http://localhost:8080/" + short))
			return

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			result["detail"] = "Method not allowed"
			return
		}
	}
}
