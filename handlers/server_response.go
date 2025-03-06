package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/egor-markin/wallcraft-go-test-task/config"
)

func writeServerResponse[T any](w http.ResponseWriter, statusCode int, data T) {
	w.Header().Set("Content-Type", config.ContentTypeJSON)
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("Error encoding server reponse: ", err)
		http.Error(w, config.InternalServerErrorMsg, http.StatusInternalServerError)
	}
}

func writeInternalServerError(w http.ResponseWriter, err error) {
	log.Println(err)
	http.Error(w, config.InternalServerErrorMsg, http.StatusInternalServerError)
}

func writeServerParseError(w http.ResponseWriter, err error) {
	log.Println(err)
	http.Error(w, "An error occurred while parsing the input JSON", http.StatusBadRequest)
}
