package app

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Group struct {
	Id          *int    `json:"id,omitempty"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	ManagerId   *int    `json:"manager_id"`
}

// -----------------------------------------------------------------------------

func getAddableMembers(w http.ResponseWriter, r *http.Request, user User) {
	if !*user.IsAdmin {
		responseMessage(w, http.StatusUnauthorized, "You cannot get this!")
		return
	}

	db, dbClose, err := openConnection()
	if err != nil {
		responseInternalError(w, err)
		return
	}
	defer dbClose()

	results, err := db.Query("SELECT `id`, `username`, `full_name` FROM `users` WHERE `group_id` IS NULL AND `is_admin` = 0")
	if err != nil {
		responseInternalError(w, err)
		return
	}

	users := make([]User, 0)

	for results.Next() {
		var user User

		err = results.Scan(&user.Id, &user.Username, &user.FullName)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		users = append(users, user)

	}

	json.NewEncoder(w).Encode(users)
}

func getMembers(w http.ResponseWriter, r *http.Request, user User) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Id must be an integer!")
		return
	}

	db, dbClose, err := openConnection()
	if err != nil {
		responseInternalError(w, err)
		return
	}
	defer dbClose()

	results, err := db.Query("SELECT `id`, `username`, `full_name` FROM `users` WHERE `group_id` = ?", id)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	users := make([]User, 0)

	for results.Next() {
		var user User

		err = results.Scan(&user.Id, &user.Username, &user.FullName)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		users = append(users, user)

	}

	json.NewEncoder(w).Encode(users)
}

func addMember(w http.ResponseWriter, r *http.Request, user User) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Id must be an integer!")
		return
	}

	db, dbClose, err := openConnection()
	if err != nil {
		responseInternalError(w, err)
		return
	}
	defer dbClose()

	idx, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Id must be an integer!")
		return
	}

	// Check addable

	_, err = db.Query("UPDATE `users` SET `group_id` = ? WHERE `id` = ?", id, idx)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	responseMessage(w, http.StatusOK, "Add member successfully!")
}

func setManager(w http.ResponseWriter, r *http.Request, user User) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Id must be an integer!")
		return
	}

	db, dbClose, err := openConnection()
	if err != nil {
		responseInternalError(w, err)
		return
	}
	defer dbClose()

	idx, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Id must be an integer!")
		return
	}

	// Check addable

	_, err = db.Query("UPDATE `groups` SET `manager_id` = ? WHERE `id` = ?", idx, id)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	responseMessage(w, http.StatusOK, "Set manager successfully!")
}

func removeMember(w http.ResponseWriter, r *http.Request, user User) {
	db, dbClose, err := openConnection()
	if err != nil {
		responseInternalError(w, err)
		return
	}
	defer dbClose()

	idx, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Id must be an integer!")
		return
	}

	// Check removable

	_, err = db.Query("UPDATE `users` SET `group_id` = NULL WHERE `id` = ?", idx)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	responseMessage(w, http.StatusOK, "Remove member successfully!")
}

func getAllGroups(w http.ResponseWriter, r *http.Request, user User) {

	db, dbClose, err := openConnection()
	if err != nil {
		responseInternalError(w, err)
		return
	}
	defer dbClose()

	results, err := db.Query("SELECT `id`, `name`, `description`, `manager_id` FROM `groups`")
	if err != nil {
		responseInternalError(w, err)
		return
	}

	groups := make([]Group, 0)

	for results.Next() {
		var group Group

		err = results.Scan(&group.Id, &group.Name, &group.Description, &group.ManagerId)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		groups = append(groups, group)

	}

	json.NewEncoder(w).Encode(groups)
}

func createGroup(w http.ResponseWriter, r *http.Request, user User) {
	if *user.IsAdmin {
		var newGroup Group
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		json.Unmarshal(reqBody, &newGroup)

		if newGroup.Name == nil {
			responseMessage(w, http.StatusBadRequest, "Group's name must not be empty!")
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
	responseMessage(w, http.StatusUnauthorized, "You are not authorized to create group!")
}

func getOneGroup(id int) (*Group, error) {

	db, dbClose, err := openConnection()
	if err != nil {
		return nil, err
	}
	defer dbClose()

	results, err := db.Query("SELECT `id`,`name`,`description`,`manager_id` FROM `groups` WHERE `id` = ?", id)
	if err != nil {
		return nil, err
	}

	var group Group

	if results.Next() {
		err = results.Scan(&group.Id, &group.Name, &group.Description, &group.ManagerId)
		if err != nil {
			return nil, err
		}
		return &group, nil
	}

	return nil, errors.New("Invalid group id")
}

func routerGetOneGroup(w http.ResponseWriter, r *http.Request, user User) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Id must be an integer!")
		return
	}

	group, err := getOneGroup(id)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	response(w, http.StatusOK, group)
}

func updateGroup(w http.ResponseWriter, r *http.Request, user User) {
	if *user.IsAdmin {
		idGr := mux.Vars(r)["id"]

		var group Group

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			responseInternalError(w, err)
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
	if *user.IsAdmin {
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
