package common

import (
	"encoding/json"
	"net/http"
)

func RespondOK(w http.ResponseWriter) {
	w.WriteHeader(200)
}

func RespondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(response))
}

func RespondError(w http.ResponseWriter, status int, err string) {
	RespondJSON(w, status, map[string]string{"error": err})
}
