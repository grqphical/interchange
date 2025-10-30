package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
)

func BuildStaticFileSystemHandler(service map[string]any, name string, route string) (http.Handler, bool) {
	dir, exists := service["directory"]
	if !exists {
		slog.Error("ConfigurationError", "err", fmt.Sprintf("directory not set in service '%s'", name))
		return nil, false
	}

	slog.Info("Serving static files", "route", route, "directory", dir)
}
