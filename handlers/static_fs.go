package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type InterchangeStaticFSHandler struct {
	route     string
	directory string
}

func (i InterchangeStaticFSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	file := strings.Replace(r.RequestURI, i.route, "", 1)
	fullFilePath := filepath.Join(i.directory, file)

	if _, err := os.Stat(fullFilePath); errors.Is(err, os.ErrNotExist) {
		fmt.Fprintf(w, "404 not found")
		return
	}

	data, err := os.ReadFile(fullFilePath)
	if err != nil {
		fmt.Fprintf(w, "404 not found")
		return
	}

	contentType := mime.TypeByExtension(filepath.Ext(fullFilePath))
	w.Header().Set("Content-Type", contentType)
	w.Write(data)
}

func BuildStaticFileSystemHandler(service map[string]any, name string, route string) (http.Handler, bool) {
	dir, exists := service["directory"]
	if !exists {
		slog.Error("ConfigurationError", "err", fmt.Sprintf("directory not set in service '%s'", name))
		return nil, false
	}

	slog.Info("Serving static files", "route", route, "directory", dir)
	directory, err := filepath.Abs(dir.(string))
	if err != nil {
		return nil, false
	}

	return InterchangeStaticFSHandler{
		route,
		directory,
	}, true
}
