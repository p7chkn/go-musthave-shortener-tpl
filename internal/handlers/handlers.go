package handlers

import "net/http"

func UrlHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("<h1>This is GET</h1>"))
	case http.MethodPost:
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("<h1>This is POST</h1>"))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("<h1>Method not allowed</h1>"))
	}

}
