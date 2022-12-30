package www

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestStaticWebsite(t *testing.T) {
	t.Parallel()

	sampleFileData := []byte("export const c = null;")
	fallbackPageData := []byte("<h1>404</h1>")

	var validStaticWebsiteConfig = StaticWebsiteConfig{
		Fsys: fstest.MapFS{
			"index.html":    {Data: []byte("<h1>Home</h1>")},
			"404.html":      {Data: fallbackPageData},
			"nested/app.js": {Data: sampleFileData},
		},
		SubDir:           ".",
		FallbackPagePath: "404.html",
	}

	t.Run("Should implement http.Handler interface", func(t *testing.T) {
		var _ http.Handler = (*StaticWebsite)(nil)
	})

	t.Run("Should serve static files from FS", func(t *testing.T) {
		h, err := NewStaticWebsite(validStaticWebsiteConfig)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/nested/app.js", nil)
		resrec := httptest.NewRecorder()
		h.ServeHTTP(resrec, req)
		result := resrec.Result()

		if result.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status code %d", result.StatusCode)
		}
		// check that body is file content
		if content, _ := ioutil.ReadAll(result.Body); !bytes.Equal(content, sampleFileData) {
			t.Fatalf("unexpected file content %v", content)
		}
	})

	t.Run("Should serve fallback page on 404", func(t *testing.T) {
		h, err := NewStaticWebsite(validStaticWebsiteConfig)
		if err != nil {
			t.Fatalf("unexpected err %v", err)
		}

		// make a request for a file that does not exist
		req := httptest.NewRequest(http.MethodGet, "/doesnt-exist", nil)
		resrec := httptest.NewRecorder()
		h.ServeHTTP(resrec, req)
		result := resrec.Result()

		if result.StatusCode != http.StatusNotFound {
			t.Fatalf("unexpected status code %d", result.StatusCode)
		}
		// check that body is fallback page content
		if content, _ := ioutil.ReadAll(result.Body); !bytes.Equal(content, fallbackPageData) {
			t.Fatalf("unexpected file content %v", content)
		}
	})
}

func TestStaticStaticWebsiteConfig(t *testing.T) {
	t.Parallel()

	var validDefaultHandler = StaticWebsiteConfig{
		Fsys: fstest.MapFS{
			"404.html":      {Data: []byte("<h1>404</h1>")},
			"index.html":    {Data: []byte("<h1>Home</h1>")},
			"nested/app.js": {Data: []byte("export const c = null;")},
		},
		SubDir:           ".",
		FallbackPagePath: "404.html",
	}

	t.Run("Should accept valid config", func(t *testing.T) {
		if err := validDefaultHandler.Validate(); err != nil {
			t.Fatalf("got unexpected err %v", err)
		}
	})

	t.Run("Should validate file system", func(t *testing.T) {
		config := validDefaultHandler // mutate copy only
		config.Fsys = nil             // add invalid fsys
		if err := config.Validate(); err == nil {
			t.Fatal("wanted an error")
		}
	})

	t.Run("Should validate sub directory", func(t *testing.T) {
		config := validDefaultHandler // copy to avoid mutation
		config.SubDir = ""            // add invalid sub dir
		if err := config.Validate(); err == nil {
			t.Fatal("wanted an error")
		}
	})

	t.Run("Should validate fallback page", func(t *testing.T) {
		config := validDefaultHandler // copy to avoid mutation
		config.FallbackPagePath = ""  // add invalid handler
		if err := config.Validate(); err == nil {
			t.Fatal("wanted an error")
		}
	})
}
