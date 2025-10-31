package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"text/template"
)

const debugHandlerTemplate string = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Interchange Debug Panel</title>
</head>
<body>
    <a href="/debug/log">Server Logs</a>
</body>
</html>
`

// handles debug requests if the server is running in developmentMode
func DebugHandler(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		w.Header().Set("Content-Type", "text/html")
		t := template.Must(template.New("debug").Parse(debugHandlerTemplate))

		t.Execute(w, nil)
		return
	}

	debugInfo := map[string]any{}

	json.NewEncoder(w).Encode(debugInfo)
}
