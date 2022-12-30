package www

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	HeaderContentType    = "Content-Type"
	ContentTypeJSON      = "application/json"
	ContentTypePlainText = "text/plain; charset=utf-8"
)

// Writes a JSON response.
func RespondJSON(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Set(HeaderContentType, ContentTypeJSON)
	return json.NewEncoder(w).Encode(v)
}

// Writes a plain text response
func RespondText(w http.ResponseWriter, status int, s string) error {
	w.WriteHeader(status)
	w.Header().Set(HeaderContentType, ContentTypePlainText)
	_, err := w.Write([]byte(s))
	return err
}

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

func NewHTTPAccessLogMiddleware(writeLogTo io.Writer) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Serve request and record response
			resrec := &Recorder{ResponseWriter: w}
			h.ServeHTTP(resrec, r)

			// Log request URL, response status code and body size
			logstr := fmt.Sprintf(
				"HTTP %d %dB %q",
				resrec.StatusCode,
				resrec.NumBodyBytesWritten,
				r.URL.Path,
			)
			w.Write([]byte(logstr))
		})
	}
}
