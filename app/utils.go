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
	log.Panicf(err.Error())
}

func responseCustomError(w http.ResponseWriter, httpCode int, message string) {
	w.WriteHeader(httpCode)
	json.NewEncoder(w).Encode(MessageResponse{
		Message: message,
	})
}

func responseOK(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}
