package http

const (
	StatusOK         = 200
	StatusCreated    = 201
	StatusBadRequest = 400
	StatusNotFound   = 404
)

func StatusText(code int) string {
	switch code {
	case StatusOK:
		return "OK"
	case StatusCreated:
		return "Created"
	case StatusBadRequest:
		return "BadRequest"
	case StatusNotFound:
		return "Not Found"
	default:
		return ""
	}
}
