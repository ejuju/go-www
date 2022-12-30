package www

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPrefixRouterConfig(t *testing.T) {
	t.Parallel()

	validBackendPrefix := "/api"
	validHTTPHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	validPrefixRouterConfig := PrefixRouterConfig{
		PathPrefix:           validBackendPrefix,
		HandlerWithoutPrefix: validHTTPHandler,
		HandlerWithPrefix:    validHTTPHandler,
	}

	t.Run("Should accept valid config", func(t *testing.T) {
		if err := validPrefixRouterConfig.Validate(); err != nil {
			t.Fatalf("got unexpected err %v", err)
		}
	})

	t.Run("Should validate backend path prefix", func(t *testing.T) {
		config := validPrefixRouterConfig // copy to avoid mutation
		config.PathPrefix = "/"           // add invalid backend path
		if err := config.Validate(); err == nil {
			t.Fatal("wanted an error")
		}
	})

	t.Run("Should validate website handler", func(t *testing.T) {
		config := validPrefixRouterConfig // copy to avoid mutation
		config.HandlerWithoutPrefix = nil // add invalid handler
		if err := config.Validate(); err == nil {
			t.Fatal("wanted an error")
		}
	})

	t.Run("Should validate backend handler", func(t *testing.T) {
		config := validPrefixRouterConfig // copy to avoid mutation
		config.HandlerWithPrefix = nil    // add invalid handler
		if err := config.Validate(); err == nil {
			t.Fatal("wanted an error")
		}
	})
}

func TestWebsiteAndBackendHTTPRouter(t *testing.T) {
	t.Parallel()

	validBackendPrefix := "/api"
	validHTTPHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	validPrefixRouterConfig := PrefixRouterConfig{
		PathPrefix:           validBackendPrefix,
		HandlerWithoutPrefix: validHTTPHandler,
		HandlerWithPrefix:    validHTTPHandler,
	}

	t.Run("Should reject invalid configuration", func(t *testing.T) {
		tests := []struct {
			config  PrefixRouterConfig
			wantErr bool
		}{
			{config: validPrefixRouterConfig, wantErr: false},
			{config: PrefixRouterConfig{}, wantErr: true},
		}

		for i, test := range tests {
			_, err := NewPrefixRouter(test.config)
			gotErr := err != nil
			if test.wantErr != gotErr {
				t.Fatalf("Want error %v but got %v at index %d", test.wantErr, gotErr, i)
			}
		}
	})

	t.Run("Should strip backend path prefix from requests", func(t *testing.T) {
		// Init backend handler without including path prefix
		backendHandler := http.NewServeMux()
		backendHandler.HandleFunc("/test0", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Declare path prefix in config
		h, err := NewPrefixRouter(PrefixRouterConfig{
			PathPrefix:           "/api",
			HandlerWithoutPrefix: validHTTPHandler,
			HandlerWithPrefix:    backendHandler,
		})
		if err != nil {
			t.Fatalf("unexpected err %v", err)
		}

		resrec := httptest.NewRecorder()
		// Include path prefix in request
		req := httptest.NewRequest(http.MethodGet, "/api/test0", nil)
		h.ServeHTTP(resrec, req)

		// Check if handler was reach (we know it is reached if we get a 200)
		if resrec.Result().StatusCode != http.StatusOK {
			t.Fatal("request did not reach handler func")
		}
	})

	t.Run("Should route requests to the right handler", func(t *testing.T) {
		websiteStatusCode := 800
		backendStatusCode := 900

		websiteHandler := http.NewServeMux()
		websiteHandler.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(websiteStatusCode)
		})

		backendHandler := http.NewServeMux()
		backendHandler.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(backendStatusCode)
		})

		h, err := NewPrefixRouter(PrefixRouterConfig{
			PathPrefix:           "/api",
			HandlerWithoutPrefix: websiteHandler,
			HandlerWithPrefix:    backendHandler,
		})
		if err != nil {
			t.Fatalf("unexpected err %v", err)
		}

		tests := []struct {
			wantStatusCode int
			url            string
		}{
			{wantStatusCode: websiteStatusCode, url: "/about"},    // request for website handler
			{wantStatusCode: backendStatusCode, url: "/api/user"}, // request for backend handler
		}

		for i, test := range tests {
			resrec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, test.url, nil)
			h.ServeHTTP(resrec, req)

			// Check if handler was reach (we know it is reached if we get the right status code)
			if resrec.Result().StatusCode != test.wantStatusCode {
				fmt.Println(resrec.Result().StatusCode, test.wantStatusCode)
				t.Fatalf("request did not reach handler func (at index: %d)", i)
			}
		}
	})
}
