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

func getManager(id int) (bool, error) {
	db, dbClose, err := openConnection()
	if err != nil {
		return false, err
	}
	defer dbClose()

	results, err := db.Query("SELECT `manager_id` FROM `groups` WHERE `id` = ?", id)
	if err != nil {
		return false, err
	}

	if results.Next() {
		var manager_id int
		err = results.Scan(&manager_id)
		if err != nil {
			return false, err
		}
	}

	return false, nil
}

func getAllGroups(w http.ResponseWriter, r *http.Request, user User) {

	db, dbClose, err := openConnection()
	if err != nil {
		responseInternalError(w, err)
		return
	}
	defer dbClose()

	results, err := db.Query("SELECT `id`, `name`, `description` FROM `groups`")
	if err != nil {
		responseInternalError(w, err)
		return
	}

	var groups []Group

	for results.Next() {
		var group Group

		err = results.Scan(&group.Id, &group.Name, &group.Description)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		groups = append(groups, group)

	}

	json.NewEncoder(w).Encode(groups)
}

func createGroup(w http.ResponseWriter, r *http.Request, user User) {
	if *user.IsAdmin == true {
		var newGroup Group
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			responseInternalError(w, err)
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

		db, dbClose, err := openConnection()
		if err != nil {
			responseInternalError(w, err)
			return
		}
		defer dbClose()

		_, err = db.Query("INSERT INTO `groups`(`name`, `description`) VALUES(?,?)", newGroup.Name, newGroup.Description)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "New group created successfully!",
		})
	}
	w.WriteHeader(http.StatusUnauthorized)
	return
}

func getOneGroup(w http.ResponseWriter, r *http.Request, user User) {
	idGr := mux.Vars(r)["id"]

	db, dbClose, err := openConnection()
	if err != nil {
		responseInternalError(w, err)
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
	} else {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(group)
	}
}

func updateGroup(w http.ResponseWriter, r *http.Request, user User) {
	if *user.IsAdmin == true {
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

		db, dbClose, err := openConnection()
		if err != nil {
			responseInternalError(w, err)
			return
		}
		defer dbClose()

		_, err = db.Query("UPDATE `groups` SET `name` = ?, `description` = ? WHERE `id` = ?", group.Name, group.Description, idGr)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Group updated!",
		})
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
}

func deleteGroup(w http.ResponseWriter, r *http.Request, user User) {
	if *user.IsAdmin == true {
		idGr := mux.Vars(r)["id"]

		db, dbClose, err := openConnection()
		if err != nil {
			responseInternalError(w, err)
			return
		}
		defer dbClose()

		_, err = db.Query("DELETE FROM `groups` WHERE `id` = ?", idGr)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Group deleted!",
		})
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
}
