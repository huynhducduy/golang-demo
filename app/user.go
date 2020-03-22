package app

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type User struct {
	Id       *int    `json:"id,omitempty"`
	Username *string `json:"username"`
	FullName *string `json:"full_name"`
	GroupId  *int    `json:"group_id"`
	IsAdmin  *bool   `json:"is_admin"`
	Password *string `json:"password,omitempty"`
}

func getOneUser(id int) (*User, error) {
	var user User

	results := db.QueryRow("SELECT `id`, `full_name`, `username`, `group_id`, `is_admin` FROM `users` WHERE `id` = ?", id)
	err := results.Scan(&user.Id, &user.FullName, &user.Username, &user.GroupId, &user.IsAdmin)
	if err == sql.ErrNoRows {
		return nil, errors.New("Invalid user id")
	} else if err != nil {
		return nil, err
	}

	return &user, nil
}

func routerGetMe(w http.ResponseWriter, r *http.Request, user User) {
	json.NewEncoder(w).Encode(user)
}

func getAllUsers(w http.ResponseWriter, r *http.Request, user User) {

	results, err := db.Query("SELECT `id`, `username`, `full_name`, `group_id`, `is_admin` FROM `users`")
	if err != nil {
		responseInternalError(w, err)
		return
	}
	defer results.Close()

	users := make([]User, 0)

	for results.Next() {
		var user User

		err = results.Scan(&user.Id, &user.Username, &user.FullName, &user.GroupId, &user.IsAdmin)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		users = append(users, user)

	}

	json.NewEncoder(w).Encode(users)
}

func createUser(w http.ResponseWriter, r *http.Request, user User) {
	if *user.IsAdmin {
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		var thisUser User

		json.Unmarshal(reqBody, &thisUser)

		if thisUser.Username == nil {
			responseMessage(w, http.StatusBadRequest, "Username must not be empty!")
			return
		}

		if thisUser.Password == nil {
			responseMessage(w, http.StatusBadRequest, "Username must not be empty!")
			return
		}

		if thisUser.FullName == nil {
			responseMessage(w, http.StatusBadRequest, "Full name must not be empty!")
			return
		}

		_, err = db.Exec("INSERT INTO `users`(`username`, `password`,`full_name`) VALUES(?,?,?)", thisUser.Username, thisUser.Password, thisUser.FullName)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		responseMessage(w, http.StatusOK, "User created!")
		return

	}
	responseMessage(w, http.StatusUnauthorized, "Unauthorized")
}

func routerGetOneUser(w http.ResponseWriter, r *http.Request, user User) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Id must be an integer!")
		return
	}
	thisUser, err := getOneUser(id)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	response(w, http.StatusOK, thisUser)
}

func updateUser(w http.ResponseWriter, r *http.Request, user User) {
	if *user.IsAdmin {
		id, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			responseMessage(w, http.StatusBadRequest, "Id must be an integer!")
			return
		}

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		var thisUser User

		json.Unmarshal(reqBody, &thisUser)

		_, err = db.Exec("UPDATE `users` SET `full_name` = ? WHERE `id` = ?", thisUser.FullName, id)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		responseMessage(w, http.StatusOK, "Update user successfully!")
		return
	}
	responseMessage(w, http.StatusUnauthorized, "Unauthorized")
}

func deleteUser(w http.ResponseWriter, r *http.Request, user User) {
	if *user.IsAdmin {
		id, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			responseMessage(w, http.StatusBadRequest, "Id must be an integer!")
			return
		}

		_, err = db.Exec("DELETE FROM `users` WHERE `id` = ?", id)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		responseMessage(w, http.StatusOK, "Delete user successfully!")
		return
	}
	responseMessage(w, http.StatusUnauthorized, "Unauthorized")
}
