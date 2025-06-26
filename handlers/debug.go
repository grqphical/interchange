package handlers

import (
	"encoding/json"
	"net/http"
)

const debugHandlerTemplate string = ``

func DebugHandler(w http.ResponseWriter, r *http.Request) {
	debugInfo := map[string]any{}

	json.NewEncoder(w).Encode(debugInfo)
}
