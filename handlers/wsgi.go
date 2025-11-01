package handlers

import (
	"bufio"
	"bytes"
	_ "embed"
	"io"
	"log/slog"
	"net/http"
	"os/exec"
	"runtime"
)

//go:embed wsgi_bridge.py
var wsgi_bridge string

func BuildWSGIHandler(service map[string]any) (http.Handler, bool) {
	var pythonCmd string
	if runtime.GOOS == "windows" {
		cmd := exec.Command("python", "--version")
		if err := cmd.Run(); err != nil {
			slog.Error("Cannot run WSGI apps without python installed")
			return nil, false
		}
		pythonCmd = "python"
	} else {
		cmd := exec.Command("python3", "--version")
		if err := cmd.Run(); err != nil {
			slog.Error("Cannot run WSGI apps without python3 installed")
			return nil, false
		}
		pythonCmd = "python3"
	}

	module, exists := service["module"]
	if !exists {
		slog.Error("ConfigurationError: module missing from WSGI service")
		return nil, false
	}

	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		cmd := exec.Command(pythonCmd, "-c", wsgi_bridge, module.(string))
		var inputBuf bytes.Buffer
		r.Write(&inputBuf)

		var outputBuf bytes.Buffer

		cmd.Stdin = &inputBuf
		cmd.Stdout = &outputBuf
		if err := cmd.Run(); err != nil {
			slog.Error("Failed to run Python WSGI bridge", "error", err.Error())
			return
		}

		resp, err := http.ReadResponse(bufio.NewReader(&outputBuf), nil)
		if err != nil {
			slog.Error("Failed to decode WSGI response", "error", err.Error())
			return
		}
		defer resp.Body.Close()

		for k, v := range resp.Header {
			for _, val := range v {
				w.Header().Add(k, val)
			}
		}

		w.WriteHeader(resp.StatusCode)

		io.Copy(w, resp.Body)
	}

	return http.HandlerFunc(handlerFunc), true
}
