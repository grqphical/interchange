package handlers

import (
	"compress/gzip"
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/grqphical/interchange/templates"
)

// a custom static file handler for interchange
type InterchangeStaticFSHandler struct {
	route        string
	directory    string
	showDirPages bool
	compression  string
}

func writeFileData(w http.ResponseWriter, data []byte, compression string) {
	switch compression {
	case "gzip":
		w.Header().Set("Content-Encoding", "gzip")
		cw := gzip.NewWriter(w)

		cw.Write(data)
		cw.Close()
	case "":
		w.Write(data)
	default:
		slog.Error("invalid compression type")
	}
}

func (i InterchangeStaticFSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	file := strings.Replace(r.RequestURI, i.route, "", 1)
	fullFilePath := filepath.Join(i.directory, file)

	info, err := os.Stat(fullFilePath)
	if errors.Is(err, os.ErrNotExist) {
		templates.WriteError(w, 404, "Not Found")
		return
	}

	data, err := os.ReadFile(fullFilePath)
	if err != nil {
		if info.IsDir() {
			// serve index.html from within the directory if it exists
			data, err := os.ReadFile(filepath.Join(fullFilePath, "index.html"))
			if err == nil {
				w.Header().Set("Content-Type", "text/html")

				writeFileData(w, data, i.compression)

				return
			} else {
				// show the directory browser if the user configured it to be shown
				if i.showDirPages {
					templates.WriteDirectoryTemplate(w, fullFilePath, r.RequestURI, i.directory)
				} else {
					templates.WriteError(w, http.StatusNotFound, "Not Found")
					return
				}
			}

		} else {
			templates.WriteError(w, 500, "Internal Server Error")
		}
		return
	}

	contentType := mime.TypeByExtension(filepath.Ext(fullFilePath))
	w.Header().Set("Content-Type", contentType)
	writeFileData(w, data, i.compression)
}

func BuildStaticFileSystemHandler(service map[string]any, name string, route string) (http.Handler, bool) {
	dir, exists := service["directory"]
	if !exists {
		slog.Error("ConfigurationError", "err", fmt.Sprintf("directory not set in service '%s'", name))
		return nil, false
	}

	directory, err := filepath.Abs(dir.(string))
	if err != nil {
		return nil, false
	}

	showDirPages, exists := service["showdirectorybrowser"]
	if !exists {
		showDirPages = true
	}
	compression, exists := service["compression"]
	if !exists {
		compression = ""
	}

	return InterchangeStaticFSHandler{
		route,
		directory,
		showDirPages.(bool),
		compression.(string),
	}, true
}
