package www

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

type StaticWebsite struct {
	fileServer   http.Handler
	fallbackHTML []byte
}

func (h *StaticWebsite) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if file server returns a 404 by passing
	// a response recorder and not the true response writer
	// If file is not found, serve fallback page
	resrec := httptest.NewRecorder()
	h.fileServer.ServeHTTP(resrec, r)
	if resrec.Result().StatusCode == http.StatusNotFound {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(h.fallbackHTML)
		return
	}

	h.fileServer.ServeHTTP(w, r)
}

type StaticWebsiteConfig struct {
	Fsys             fs.FS  // static file system
	SubDir           string // "." for current directory
	FallbackPagePath string // ex: "404.html"
}

func (c StaticWebsiteConfig) Validate() error {
	switch {
	default:
		return nil
	case c.Fsys == nil:
		return errors.New("no file system was provided")
	case c.SubDir == "":
		return errors.New("sub directory path is empty")
	case c.FallbackPagePath == "":
		return errors.New("fallback page path is empty")
	}
}

// subFS: Set it to "." for current dir
func NewStaticWebsite(c StaticWebsiteConfig) (*StaticWebsite, error) {
	// Get sub directory where static files are
	websiteFS, err := fs.Sub(c.Fsys, c.SubDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get website build sub file-system: %w", err)
	}

	// Get fallback page content
	f, err := websiteFS.Open(c.FallbackPagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open fallback page: %w", err)
	}
	defer f.Close()
	fallbackPageData, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read fallback page content: %w", err)
	}

	return &StaticWebsite{
		fileServer:   http.FileServer(http.FS(websiteFS)),
		fallbackHTML: fallbackPageData,
	}, nil
}

func (sw *StaticWebsite) HTTPHandler() (string, http.Handler) {
	return "/", sw
}
