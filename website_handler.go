package httputils

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

// Handles static websites
type WebsiteHandler struct {
	fsys         fs.FS
	fileServer   http.Handler
	fallbackHTML []byte
}

func (h *WebsiteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if file server returns a 404 by passing
	// a response recorder and not the true response writer
	resrec := httptest.NewRecorder()
	h.fileServer.ServeHTTP(resrec, r)

	// If file is not found, serve fallback page
	if resrec.Result().StatusCode == http.StatusNotFound {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(h.fallbackHTML))
		return
	}

	h.fileServer.ServeHTTP(w, r)
}

type WebsiteHandlerConfig struct {
	Fsys             fs.FS  // static file system
	SubDir           string // "." for current directory
	FallbackPagePath string // ex: "404.html"
}

func (c WebsiteHandlerConfig) Validate() error {
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
func NewWebsiteHandler(c WebsiteHandlerConfig) (http.Handler, error) {
	// Get sub directory where static files are
	websiteFS, err := fs.Sub(c.Fsys, c.SubDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get website build sub file-system: %w", err)
	}

	// debug: print file paths in fs
	// p, err := paths(websiteFS)
	// fmt.Printf("Printing website FS:\nErr: %v\nPaths: %v\n", err, fmtPaths(p))

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

	return &WebsiteHandler{
		fileServer:   http.FileServer(http.FS(websiteFS)),
		fsys:         websiteFS,
		fallbackHTML: fallbackPageData,
	}, nil
}

// // Debug utils
// func fmtPaths(paths map[string]int) string {
// 	outputLines := []string{}
// 	totalSize := 0
// 	i := 1
// 	for filepath, size := range paths {
// 		fileKB := float64(size) / 1_000.0
// 		line := fmt.Sprintf("%8.3f KB  %s\n", fileKB, filepath)
// 		outputLines = append(outputLines, line)
// 		totalSize += size
// 		i++
// 	}

// 	// put html files first
// 	sort.Slice(outputLines, func(i, j int) bool {
// 		return strings.Contains(outputLines[i], ".html")
// 	})

// 	output := fmt.Sprintf("Website FS has %d files (total size: %.3f MB)\n", len(paths), float64(totalSize)/1_000_000.0)
// 	for i, line := range outputLines {
// 		output += fmt.Sprintf("\t%3d. %s", i+1, line)
// 	}
// 	return output
// }

// func paths(fsys fs.FS) (map[string]int, error) {
// 	paths := map[string]int{}
// 	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
// 		if err != nil {
// 			return err
// 		}
// 		if d.IsDir() {
// 			return nil
// 		}
// 		info, err := d.Info()
// 		if err != nil {
// 			return err
// 		}
// 		paths[path] = int(info.Size())
// 		return nil
// 	})
// 	return paths, err
// }
