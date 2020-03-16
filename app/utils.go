package app

import (
	"log"
	"net/http"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type MessageResponse struct {
	Message string `json:"message"`
}

func responseInternalError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(MessageResponse{
		Message: "Internal error!",
	})
	log.Printf(err.Error())
}

func responseCustomError(w http.ResponseWriter, err error, httpCode int, message string) {
	w.WriteHeader(httpCode)
	json.NewEncoder(w).Encode(MessageResponse{
		Message: message,
	})
	log.Printf(err.Error())
}
