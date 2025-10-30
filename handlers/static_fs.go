package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
)

type InterchangeStaticFSHandler struct{}

func (i InterchangeStaticFSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "url: %s\n", r.RequestURI)
}

func BuildStaticFileSystemHandler(service map[string]any, name string, route string) (http.Handler, bool) {
	dir, exists := service["directory"]
	if !exists {
		slog.Error("ConfigurationError", "err", fmt.Sprintf("directory not set in service '%s'", name))
		return nil, false
	}

	slog.Info("Serving static files", "route", route, "directory", dir)

	return InterchangeStaticFSHandler{}, true
}
