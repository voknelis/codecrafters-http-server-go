package handlers

import "github.com/codecrafters-io/http-server-starter-go/app/http"

func EchoHandler(w http.ResponseWriter, r *http.Request) error {
	echoValue := r.PathValue("value")

	w.StatusCode = http.StatusOK
	w.Headers["Content-Type"] = "text/plain"

	_, err := w.Write([]byte(echoValue))
	return err
}
