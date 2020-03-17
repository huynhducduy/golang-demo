package app

import (
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
	db, dbClose, err := openConnection()
	if err != nil {
		return nil, err
	}
	defer dbClose()

	results, err := db.Query("SELECT `id`, `full_name`, `username`, `group_id`, `is_admin` FROM `users` WHERE `id` = ?", id)
	if err != nil {
		return nil, err
	}

	if results.Next() {
		var user User

		err = results.Scan(&user.Id, &user.FullName, &user.Username, &user.GroupId, &user.IsAdmin)
		if err != nil {
			return nil, err
		}

		return &user, nil
	} else {
		return nil, errors.New("Invalid user id")
	}
}

func routerGetMe(w http.ResponseWriter, r *http.Request, user User) {
	json.NewEncoder(w).Encode(user)
}

func getAllUsers(w http.ResponseWriter, r *http.Request, user User) {

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

		db, dbClose, err := openConnection()
		if err != nil {
			responseInternalError(w, err)
			return
		}
		defer dbClose()

		_, err = db.Query("INSERT INTO `users`(`username`, `password`,`full_name`) VALUES(?,?,?)", thisUser.Username, thisUser.Password, thisUser.FullName)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		responseMessage(w, http.StatusOK, "User created!")

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

		db, dbClose, err := openConnection()
		if err != nil {
			responseInternalError(w, err)
			return
		}
		defer dbClose()

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		var thisUser User

		json.Unmarshal(reqBody, &thisUser)

		_, err = db.Query("UPDATE `users` SET `full_name` = ? WHERE `id` = ?", thisUser.FullName, id)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		responseMessage(w, http.StatusOK, "Delete user successfully!")

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

		db, dbClose, err := openConnection()
		if err != nil {
			responseInternalError(w, err)
			return
		}
		defer dbClose()

		_, err = db.Query("DELETE FROM `users` WHERE `id` = ?", id)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		responseMessage(w, http.StatusOK, "Delete user successfully!")
	}
	responseMessage(w, http.StatusUnauthorized, "Unauthorized")
}

// func isManagerOf(groupId int, userId int) (bool, error) {

// 	db, dbClose, err := openConnection()
// 	if err != nil {
// 		return false, err
// 	}
// 	defer dbClose()

// 	results, err := db.Query("SELECT `manager_id` FROM `groups` WHERE `id` = ", groupId)
// 	if err != nil {
// 		log.Printf(err.Error())
// 		return false, nil
// 	}

// 	if results.Next() {

// 		var manager_id int

// 		results.Scan(&manager_id)

// 		if manager_id == userId {
// 			return true, nil
// 		} else {
// 			return false, nil
// 		}
// 	}

// 	return false, nil
// }
