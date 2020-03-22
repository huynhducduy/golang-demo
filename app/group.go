package app

import (
	"database/sql"
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

	results, err := db.Query("SELECT `id`, `username`, `full_name` FROM `users` WHERE `group_id` IS NULL AND `is_admin` = 0")
	if err != nil {
		responseInternalError(w, err)
		return
	}
	defer results.Close()

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

	results, err := db.Query("SELECT `id`, `username`, `full_name`, `group_id` FROM `users` WHERE `group_id` = ?", id)
	if err != nil {
		responseInternalError(w, err)
		return
	}
	defer results.Close()

	users := make([]User, 0)

	for results.Next() {
		var user User

		err = results.Scan(&user.Id, &user.Username, &user.FullName, &user.GroupId)
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

	idx, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Id must be an integer!")
		return
	}

	// Check addable

	_, err = db.Exec("UPDATE `users` SET `group_id` = ? WHERE `id` = ?", id, idx)
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

	idx, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Id must be an integer!")
		return
	}

	// Check addable

	_, err = db.Exec("UPDATE `groups` SET `manager_id` = ? WHERE `id` = ?", idx, id)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	responseMessage(w, http.StatusOK, "Set manager successfully!")
}

func removeMember(w http.ResponseWriter, r *http.Request, user User) {

	idx, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Id must be an integer!")
		return
	}

	// Check removable

	_, err = db.Exec("UPDATE `users` SET `group_id` = NULL WHERE `id` = ?", idx)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	responseMessage(w, http.StatusOK, "Remove member successfully!")
}

func getAllGroups(w http.ResponseWriter, r *http.Request, user User) {

	results, err := db.Query("SELECT `id`, `name`, `description`, `manager_id` FROM `groups`")
	if err != nil {
		responseInternalError(w, err)
		return
	}
	defer results.Close()

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

		results, err := db.Exec("INSERT INTO `groups`(`name`, `description`) VALUES(?,?)", newGroup.Name, newGroup.Description)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		lid, err := results.LastInsertId()
		if err != nil {
			responseInternalError(w, err)
			return
		}

		responseCreated(w, lid)
		return
	}
	responseMessage(w, http.StatusUnauthorized, "You are not authorized to create group!")
}

func getOneGroup(id int) (*Group, error) {
	var group Group

	results := db.QueryRow("SELECT `id`,`name`,`description`,`manager_id` FROM `groups` WHERE `id` = ?", id)
	err := results.Scan(&group.Id, &group.Name, &group.Description, &group.ManagerId)
	if err == sql.ErrNoRows {
		return nil, errors.New("Invalid group id")
	} else if err != nil {
		return nil, err
	}

	return &group, nil
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
			responseMessage(w, http.StatusBadRequest, "Group's name must not be empty!")
			return
		}

		_, err = db.Exec("UPDATE `groups` SET `name` = ?, `description` = ?, `manager_id` = ? WHERE `id` = ?", group.Name, group.Description, group.ManagerId, idGr)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		responseMessage(w, http.StatusOK, "Group updated!")
		return
	}
	w.WriteHeader(http.StatusUnauthorized)
}

func deleteGroup(w http.ResponseWriter, r *http.Request, user User) {
	if *user.IsAdmin {
		idGr := mux.Vars(r)["id"]

		_, err := db.Exec("DELETE FROM `groups` WHERE `id` = ?", idGr)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		responseMessage(w, http.StatusOK, "Group deleted!")
	}
	w.WriteHeader(http.StatusUnauthorized)
}
