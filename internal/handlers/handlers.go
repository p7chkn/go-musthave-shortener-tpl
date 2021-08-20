package handlers

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/p7chkn/go-musthave-shortener-tpl/internal/shortener"
)

func URLHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	f := func(c rune) bool {
		return c == '/'
	}

	splitURL := strings.FieldsFunc(r.URL.Path, f)

	if len(splitURL) > 1 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Page not found"))
		return
	}

	switch r.Method {

	case http.MethodGet:

		if len(splitURL) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad request"))
			return
		}

		long, err := shortener.GetURL(splitURL[0])

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not found"))
			return
		}
		w.Header().Set("Location", long)
		w.WriteHeader(http.StatusTemporaryRedirect)

	case http.MethodPost:

		if len(splitURL) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad request"))
			return
		}

		defer r.Body.Close()

		body, err := ioutil.ReadAll(r.Body)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad request"))
			return
		}

		short := shortener.AddURL(string(body))
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(short))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
	}
}
