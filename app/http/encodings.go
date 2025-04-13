package http

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"slices"
	"strings"
)

var acceptEncoding = []string{"gzip"}

type Encodings struct {
	encodings []string
}

func (e *Encodings) Parse(rawEncoding string) {
	encodings := strings.Split(rawEncoding, ",")
	for _, v := range encodings {
		v = strings.TrimSpace(v)
		if slices.Contains(acceptEncoding, v) {
			e.encodings = append(e.encodings, v)
		}
	}
}

func (e *Encodings) GetEncoding() string {
	if len(e.encodings) > 0 {
		return e.encodings[0]
	}
	return ""
}

func GetEncodedContent(encoding, content string) (string, error) {
	if encoding == "gzip" {
		var zbuf bytes.Buffer
		zw := gzip.NewWriter(&zbuf)

		_, err := zw.Write([]byte(content))
		if err != nil {
			return "", err
		}

		err = zw.Close()
		if err != nil {
			return "", err
		}

		return zbuf.String(), nil
	}

	return content, nil
}

func GetDecodedContent(encoding, content string) (string, error) {
	if encoding == "gzip" {
		var zbuf bytes.Buffer
		zr, err := gzip.NewReader(&zbuf)
		if err != nil {
			return "", err
		}

		_, err = zr.Read([]byte(content))
		if err != nil {
			return "", err
		}

		err = zr.Close()
		if err != nil {
			return "", err
		}

		return zbuf.String(), nil
	}

	return "", fmt.Errorf("invalid encoding: %s", encoding)
}
