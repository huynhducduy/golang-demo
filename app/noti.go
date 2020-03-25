package app

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Noti struct {
	Id      *int    `json:"id,omitempty"`
	Message *string `json:"message"`
	TaskId  *int    `json:"task_id"`
	UserId  *int    `json:"user_id"`
	Read    *bool   `json:"read"`
}

func getAllNotis(w http.ResponseWriter, r *http.Request, user User) {

	results, err := db.Query("SELECT `id`, `message`, `task_id`, `user_id`, `read` FROM `notifications` WHERE `user_id` = ? ORDER BY `id` DESC", *user.Id)
	if err != nil {
		responseInternalError(w, err)
		return
	}
	defer results.Close()

	notis := make([]Noti, 0)

	for results.Next() {
		var noti Noti

		err = results.Scan(&noti.Id, &noti.Message, &noti.TaskId, &noti.UserId, &noti.Read)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		notis = append(notis, noti)

	}

	json.NewEncoder(w).Encode(notis)
}

func readAllNotis(w http.ResponseWriter, r *http.Request, user User) {
	_, err := db.Exec("UPDATE `notifications` SET `read` = 1 WHERE `user_id` = ?", user.Id)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	responseMessage(w, http.StatusOK, "OK!")
}

func readNoti(w http.ResponseWriter, r *http.Request, user User) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Id must be an integer!")
		return
	}

	_, err = db.Exec("UPDATE `notifications` SET `read` = 1 WHERE `user_id` = ? AND `id` = ?", user.Id, id)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	responseMessage(w, http.StatusOK, "OK!")
}

func saveToken(w http.ResponseWriter, r *http.Request, user User) {

	_, err := db.Exec("DELETE FROM `token` WHERE `token` =  ?", r.URL.Query().Get("token"))
	if err != nil {
		responseInternalError(w, err)
		return
	}

	_, err = db.Exec("INSERT INTO `token`(`user_id`, `token`) VALUES(?,?)", *user.Id, r.URL.Query().Get("token"))
	if err != nil {
		responseInternalError(w, err)
		return
	}
}
