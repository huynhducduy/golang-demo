package app

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Group struct {
	Id          *string `json:"id,omitempty"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func getAllGroups(w http.ResponseWriter, r *http.Request, id int) {

	db, dbClose := openConnection()
	if db == nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		return
	}
	defer dbClose()

	results, err := db.Query("SELECT `id`, `name`, `description` FROM `groups`")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		log.Printf(err.Error())
		return
	}

	var groups []Group

	for results.Next() {
		var group Group

		err = results.Scan(&group.Id, &group.Name, &group.Description)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(MessageResponse{
				Message: "Internal error!",
			})
			log.Printf(err.Error())
			return
		}

		groups = append(groups, group)

	}

	json.NewEncoder(w).Encode(groups)
}

func createGroup(w http.ResponseWriter, r *http.Request, id int) {
	var newGroup Group
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		log.Printf(err.Error())
		return
	}

	json.Unmarshal(reqBody, &newGroup)

	if newGroup.Name == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Group's name must not be empty!",
		})
		return
	}

	db, dbClose := openConnection()
	if db == nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		return
	}
	defer dbClose()

	_, err = db.Query("INSERT INTO `groups`(`name`, `description`) VALUES(?,?)", newGroup.Name, newGroup.Description)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		log.Printf(err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(MessageResponse{
		Message: "New group created successfully!",
	})
}

func getOneGroup(w http.ResponseWriter, r *http.Request, id int) {
	idGr := mux.Vars(r)["id"]

	db, dbClose := openConnection()
	if db == nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		return
	}
	defer dbClose()

	results, err := db.Query("SELECT `id`,`name`,`description` FROM `groups` WHERE `id` = ?", idGr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		log.Printf(err.Error())
		return
	}

	var group Group

	results.Next()

	err = results.Scan(&group.Id, &group.Name, &group.Description)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		log.Printf(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(group)
}

func updateGroup(w http.ResponseWriter, r *http.Request, id int) {
	idGr := mux.Vars(r)["id"]

	var group Group

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		log.Printf(err.Error())
		return
	}

	json.Unmarshal(reqBody, &group)

	if group.Name == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Group's name must not be empty!",
		})
		return
	}

	db, dbClose := openConnection()
	if db == nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		return
	}
	defer dbClose()

	_, err = db.Query("UPDATE `groups` SET `name` = ?, `description` = ? WHERE `id` = ?", group.Name, group.Description, idGr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		log.Printf(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(MessageResponse{
		Message: "Group updated!",
	})
}

func deleteGroup(w http.ResponseWriter, r *http.Request, id int) {
	idGr := mux.Vars(r)["id"]

	db, dbClose := openConnection()
	if db == nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		return
	}
	defer dbClose()

	_, err := db.Query("DELETE FROM `groups` WHERE `id` = ?", idGr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		log.Printf(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(MessageResponse{
		Message: "Group deleted!",
	})
}
