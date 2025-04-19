package handlers

import (
	"os"
	"path/filepath"

	"github.com/codecrafters-io/http-server-starter-go/app/http"
)

func getFilesDir() string {
	args := os.Args

	if len(args) <= 2 {
		return "/tmp"
	}
	return args[2]
}

func GetFileHandler(w http.ResponseWriter, r *http.Request) error {
	dir := getFilesDir()

	filename := r.PathValue("filename")
	filePath := filepath.Join(dir, filename)

	file, err := os.ReadFile(filePath)

	if err != nil {
		w.StatusCode = http.StatusNotFound
	} else {
		w.StatusCode = http.StatusOK
		w.Headers["Content-Type"] = "application/octet-stream"
	}

	_, err = w.Write(file)
	return err
}

func CreateFileHandler(w http.ResponseWriter, r *http.Request) error {
	dir := os.Args[2]
	filename := r.PathValue("filename")
	filePath := filepath.Join(dir, filename)

	data := []byte(r.Body)
	err := os.WriteFile(filePath, data, 0777)

	body := []byte{}

	if err != nil {
		w.StatusCode = http.StatusBadRequest
		body = []byte(err.Error())
	} else {
		w.StatusCode = http.StatusCreated
	}

	_, err = w.Write(body)
	return err
}
