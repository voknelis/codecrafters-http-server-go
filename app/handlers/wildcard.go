package handlers

import "github.com/codecrafters-io/http-server-starter-go/app/http"

func WildcardHandler(w http.ResponseWriter, r *http.Request) error {
	w.StatusCode = http.StatusOK

	_, err := w.Write([]byte{})
	return err
}
