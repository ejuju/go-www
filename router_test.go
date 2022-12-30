package httputils

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGorillaMuxRouter(t *testing.T) {
	t.Parallel()

	t.Run("Implements the Router interface", func(t *testing.T) {
		var _ Router = &GorillaRouter{}
	})

	t.Run("Can register a http.HandlerFunc with an endpoint URI and method", func(t *testing.T) {
		h := NewGorillaRouter()
		endpointURI := "/test314"
		endpointMethod := http.MethodPost
		h.HandleEndpoint(endpointURI, http.MethodPost, returnStatusOK)

		req := httptest.NewRequest(endpointMethod, endpointURI, nil)
		got := serveTestHTTP(h, req)
		if got.StatusCode != http.StatusOK {
			t.Fatalf("want 200 status code but got %d", got.StatusCode)
		}
	})

	t.Run("Can register a http.Handler for a path", func(t *testing.T) {
		h := NewGorillaRouter()
		pathPrefix := "/prefix314"
		h.HandlePrefix(pathPrefix, http.HandlerFunc(returnStatusOK))

		// Ensure a request to sub path is handled
		req := httptest.NewRequest(http.MethodGet, pathPrefix+"/subpath/banana", nil)
		got := serveTestHTTP(h, req)
		if got.StatusCode != http.StatusOK {
			t.Fatalf("want 200 status code but got %d", got.StatusCode)
		}

		// Ensure a request to endpoint outside of sub path is not handled
		req = httptest.NewRequest(http.MethodGet, "/outside", nil)
		got = serveTestHTTP(h, req)
		if got.StatusCode != http.StatusNotFound {
			t.Fatalf("want 404 status code but got %d", got.StatusCode)
		}
	})
}

// sample HTTP handler func that simply returns a 200
func returnStatusOK(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func serveTestHTTP(h http.Handler, r *http.Request) *http.Response {
	resrec := httptest.NewRecorder()
	h.ServeHTTP(resrec, r)
	return resrec.Result()
}
