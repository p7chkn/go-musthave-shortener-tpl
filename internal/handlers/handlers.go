package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/p7chkn/go-musthave-shortener-tpl/internal/shortener"
)

func UrlHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	result := map[string]string{}

	f := func(c rune) bool {
		return c == '/'
	}

	splitURL := strings.FieldsFunc(r.URL.Path, f)

	if len(splitURL) > 1 {
		w.WriteHeader(http.StatusNotFound)
		result["detail"] = "Page not found"
		json.NewEncoder(w).Encode(result)
		return
	}

	switch r.Method {

	case http.MethodGet:

		if len(splitURL) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			result["detail"] = "Bad request"
			json.NewEncoder(w).Encode(result)
			return
		}

		long, err := shortener.GetUrl(splitURL[0])

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			result["detail"] = "Not found"
			json.NewEncoder(w).Encode(result)
			return
		}
		w.Header().Set("Location", long)
		w.WriteHeader(http.StatusTemporaryRedirect)

	case http.MethodPost:

		if len(splitURL) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			result["detail"] = "Bad request"
			json.NewEncoder(w).Encode(result)
			return
		}

		defer r.Body.Close()

		body, err := ioutil.ReadAll(r.Body)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			result["detail"] = "Bad request"
			json.NewEncoder(w).Encode(result)
			return
		}

		short := shortener.AddUrl(string(body))
		w.WriteHeader(http.StatusCreated)
		result["result"] = short
		json.NewEncoder(w).Encode(result)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		result["detail"] = "Method not allowed"
		json.NewEncoder(w).Encode(result)
	}

}
