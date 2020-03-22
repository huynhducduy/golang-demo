package app

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type MessageResponse struct {
	Message string `json:"message"`
}

type CreatedResponse struct {
	Id int64 `json:"id"`
}

func responseInternalError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(MessageResponse{
		Message: "Internal error!",
	})
	log.Printf(err.Error() + "\n" + string(debug.Stack()))
}

func responseMessage(w http.ResponseWriter, httpCode int, message string) {
	w.WriteHeader(httpCode)
	json.NewEncoder(w).Encode(MessageResponse{
		Message: message,
	})
}

func responseCreated(w http.ResponseWriter, id int64) {
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreatedResponse{
		Id: id,
	})
}

func response(w http.ResponseWriter, httpCode int, data interface{}) {
	w.WriteHeader(httpCode)
	json.NewEncoder(w).Encode(data)
}

func logg(x interface{}) {
	fmt.Printf("%+v\n", x)
}
