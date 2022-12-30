package httputils

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"
)

// PrefixRouter handles requests for a website and a backend API.
type PrefixRouter struct {
	config PrefixRouterConfig
}

type PrefixRouterConfig struct {
	PathPrefix           string
	HandlerWithPrefix    http.Handler
	HandlerWithoutPrefix http.Handler
}

func (c PrefixRouterConfig) Validate() error {
	switch {
	default:
		return nil
	case c.HandlerWithoutPrefix == nil:
		return errors.New("website handler is nil")
	case c.HandlerWithPrefix == nil:
		return errors.New("backend handler is nil")
	case c.PathPrefix == "":
		return errors.New("backend path prefix is empty")
	case !strings.HasPrefix(c.PathPrefix, "/"):
		return errors.New("backend path prefix should start with a slash")
	case utf8.RuneCountInString(c.PathPrefix) < 2:
		return errors.New("backend path prefix include a character after the slash")
	}
}

func NewPrefixRouter(config PrefixRouterConfig) (*PrefixRouter, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Strip prefix from backend handler request paths
	config.HandlerWithPrefix = http.StripPrefix(config.PathPrefix, config.HandlerWithPrefix)

	return &PrefixRouter{config: config}, nil
}

// This function routes requests to the appropriate handler
// depending if they are for the backend API or the file server (= website files).
func (h *PrefixRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, h.config.PathPrefix) {
		h.config.HandlerWithPrefix.ServeHTTP(w, r)
		return
	}

	h.config.HandlerWithoutPrefix.ServeHTTP(w, r)
}
