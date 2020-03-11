package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// type Credentials struct {
// 	Username string `json:"username"`
// 	Password string `json:"password"`
// }

type allGroups []Group

var groups = allGroups{
	{
		Id:          "1",
		Name:        "Introduction to Golang",
		Description: "Come join us for a chance to learn how golang works and get to eventually try it out",
	},
}

type Group struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome")
}

func login(w http.ResponseWriter, r *http.Request) {
}

func register(w http.ResponseWriter, r *http.Request) {
}

func getAllGroups(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(groups)
}

func createGroup(w http.ResponseWriter, r *http.Request) {
	var newEvent Group
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
	}

	json.Unmarshal(reqBody, &newEvent)
	groups = append(groups, newEvent)
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(newEvent)
}

func getOneGroup(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	for _, group := range groups {
		if group.Id == id {
			json.NewEncoder(w).Encode(group)
		}
	}
}

func updateGroup(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var updatedGroup Group

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
	}
	json.Unmarshal(reqBody, &updatedGroup)

	for i, group := range groups {
		if group.Id == id {
			group.Name = updatedGroup.Name
			group.Description = updatedGroup.Description
			groups = append(groups[:i], group)
			json.NewEncoder(w).Encode(group)
		}
	}
}

func deleteGroup(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	for i, group := range groups {
		if group.Id == id {
			groups = append(groups[:i], groups[i+1:]...)
			fmt.Fprintf(w, "The event with ID %v has been deleted successfully", id)
		}
	}
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/", home)
	router.HandleFunc("/api/v1/auth/login", login).Methods("POST")
	router.HandleFunc("/api/v1/auth/register", register).Methods("POST")

	router.HandleFunc("/api/v1/group", getAllGroups).Methods("GET")
	router.HandleFunc("/api/v1/group", createGroup).Methods("POST")
	router.HandleFunc("/api/v1/group/{id}", getOneGroup).Methods("GET")
	router.HandleFunc("/api/v1/group/{id}", updateGroup).Methods("PATCH")
	router.HandleFunc("/api/v1/group/{id}", deleteGroup).Methods("DELETE")

	fmt.Println("Running at port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
