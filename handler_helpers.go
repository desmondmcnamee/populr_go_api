package main

import (
	"encoding/json"
	"net/http"
)

func Respond(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	if obj, ok := data.(Public); ok {
		data = obj.Public()
	}

	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(&Resource{Data: data})
}