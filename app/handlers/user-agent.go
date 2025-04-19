package handlers

import "github.com/codecrafters-io/http-server-starter-go/app/http"

func UserAgentHandler(w http.ResponseWriter, r *http.Request) error {
	userAgent := r.Headers["User-Agent"]

	w.Headers["Content-Type"] = "text/plain"

	_, err := w.Write([]byte(userAgent))
	return err
}
