package httputils

import (
	"net/http"
)

// Recorder implements the http.ResponseWriter interface.
// It stores data about the current request / response and forwards it
// to the underlying response writer
type Recorder struct {
	http.ResponseWriter
	StatusCode          int
	NumBodyBytesWritten int
}

func (w *Recorder) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *Recorder) Write(data []byte) (int, error) {
	num, err := w.ResponseWriter.Write(data)
	w.NumBodyBytesWritten = num
	return num, err
}
