package app

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type Task struct {
	Id          *int    `json:"id,omitempty"`
	Name        *string `json:"name"`        // Updatable
	Description *string `json:"description"` // Updatable
	Report      *string `json:"report"`      // Updatable
	Assigner    *int    `json:"assigner"`
	Assignee    *int    `json:"assignee"`
	Review      *int    `json:"review"`
	ReviewAt    *int64  `json:"review_at"`
	Comment     *string `json:"comment"`
	Proof       *string `json:"proof"`
	StartAt     *int64  `json:"start_at"`
	StopAt      *int64  `json:"stop_at"`
	CloseAt     *int64  `json:"close_at"`
	OpenAt      *int64  `json:"open_at"`
	OpenFrom    *int    `json:"open_from"`
	Status      *int    `json:"status"`
	IsClosed    *bool   `json:"is_closed"`
}

func getOneTask(id int) (*Task, error) {

	results, err := db.Query("SELECT `id`, `name`, `description`,`report`,`assigner`,`assignee`,`review`,`review_at`,`comment`,`proof`,`start_at`,`close_at`,`open_at`,`open_from`,`status`,`is_closed` FROM `tasks` WHERE `id` = ?", id)
	if err != nil {
		return nil, err
	}

	if results.Next() {
		var task Task

		err = results.Scan(&task.Id, &task.Name, &task.Description, &task.Report, &task.Assigner, &task.Assignee, &task.Review, &task.ReviewAt, &task.Comment, &task.Proof, &task.StartAt, &task.OpenAt, &task.OpenAt, &task.OpenFrom, &task.Status, &task.IsClosed)
		if err != nil {
			return nil, err
		}

		return &task, nil
	} else {
		return nil, errors.New("Invalid group id")
	}

}

// -----------------------------------------------------------------------------

func getReopenableTasks(w http.ResponseWriter, r *http.Request, user User) {

	query := "SELECT `id`, `name` FROM `tasks` WHERE (`is_closed` = TRUE AND `status` != 4) OR `status` != 5"

	var mng int
	var results *sql.Rows
	var err error

	if !*user.IsAdmin {
		if user.GroupId != nil {
			group, err := getOneGroup(*user.GroupId)
			if err != nil {
				responseInternalError(w, err)
				return
			}

			if *group.ManagerId == *user.Id {
				results, err = db.Query(query+" `assigner` = ?", mng)
				if err != nil {
					responseInternalError(w, err)
					return
				}
			}
		}
	} else {
		results, err = db.Query(query)
		if err != nil {
			responseInternalError(w, err)
			return
		}
	}

	tasks := make([]Task, 0)

	for results.Next() {
		var task Task

		err = results.Scan(&task.Id, &task.Name)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		tasks = append(tasks, task)

	}

	json.NewEncoder(w).Encode(tasks)
}

func getAssignableUsers(w http.ResponseWriter, r *http.Request, user User) {

	var results *sql.Rows
	var err error

	query := "SELECT `id`, `full_name`, `username`, `group_id` FROM `users` WHERE `is_admin` = 0 AND `id` != ?"

	if *user.IsAdmin {
		results, err = db.Query(query, user.Id)
	} else {
		results, err = db.Query(query+" `group_id` = ?", user.Id, user.GroupId)
	}
	if err != nil {
		responseInternalError(w, err)
		return
	}

	users := make([]User, 0)

	for results.Next() {
		var user User

		err = results.Scan(&user.Id, &user.FullName, &user.Username, &user.GroupId)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		users = append(users, user)

	}

	json.NewEncoder(w).Encode(users)
}

func checkTask(w http.ResponseWriter, r *http.Request, user User) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Id must be an integer!")
		return
	}

	task, err := getOneTask(id)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	if !*user.IsAdmin {
		if *task.Assigner != *user.Id {
			responseMessage(w, http.StatusUnauthorized, "You cannot approve this task!")
			return
		}
	}

	if *task.Status != 0 || *task.IsClosed {
		responseMessage(w, http.StatusBadRequest, "You can only approve new tasks!")
		return
	}

	attr := "is_closed"
	if r.URL.Query().Get("close") != "true" {
		attr = "status"
	}

	_, err = db.Query("UPDATE `tasks` SET `"+attr+"` = 1 WHERE `id` = ?", id)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	responseMessage(w, http.StatusOK, "Task checked!")
}

func confirmTask(w http.ResponseWriter, r *http.Request, user User) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Id must be an integer!")
		return
	}

	task, err := getOneTask(id)
	if err != nil {
		response(w, http.StatusNotFound, err.Error())
		return
	}

	if *task.Status != 2 || *task.IsClosed || *task.Assignee != *user.Id {
		responseMessage(w, http.StatusBadRequest, "Cannot confirm this task!")
		return
	}

	r.ParseMultipartForm(10 << 20)

	file, _, err := r.FormFile("proof")
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Please uplaod proof!")
		return
	}

	defer file.Close()

	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZÅÄÖ" +
		"abcdefghijklmnopqrstuvwxyzåäö" +
		"0123456789")
	length := 8
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}

	filename := strconv.FormatInt(time.Now().Unix(), 10) + "-" + b.String() + ".png"

	uploadFile, err := os.OpenFile("images/"+filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
	}
	defer uploadFile.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}

	uploadFile.Write(fileBytes)

	stt := 3
	if r.URL.Query().Get("blocked") == "true" {
		stt = 4
	}
	_, err = db.Query("UPDATE `tasks` SET `status` = ?, `proof` = ? WHERE `id` = ?", stt, filename, id)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	responseMessage(w, http.StatusOK, "Confirm task successfully!")
}

func verifyTask(w http.ResponseWriter, r *http.Request, user User) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Id must be an integer!")
		return
	}

	task, err := getOneTask(id)
	if err != nil {
		response(w, http.StatusNotFound, err.Error())
		return
	}

	if (*task.Status != 3 && *task.Status != 4) || *task.IsClosed || (*task.Assigner != *user.Id && !*user.IsAdmin) {
		responseMessage(w, http.StatusBadRequest, "Cannot confirm this task!")
		return
	}

	if r.URL.Query().Get("ok") == "false" {
		_, err = db.Query("UPDATE `tasks` SET `status` = 2 `id` = ?", id)
	} else {
		_, err = db.Query("UPDATE `tasks` SET `is_closed` = TRUE `id` = ?", id)
	}

	if err != nil {
		responseInternalError(w, err)
		return
	}

	responseMessage(w, http.StatusOK, "Verify task successfully")
}

// -----------------------------------------------------------------------------

func getAllTasks(w http.ResponseWriter, r *http.Request, user User) {

	results, err := db.Query("SELECT `id`, `name`, `description`,`report`,`assigner`,`assignee`,`review`,`review_at`,`comment`,`proof`,`start_at`,`close_at`,`open_at`,`open_from`,`status`,`is_closed` FROM `tasks`")
	if err != nil {
		responseInternalError(w, err)
		return
	}

	tasks := make([]Task, 0)

	for results.Next() {
		var task Task

		err = results.Scan(&task.Id, &task.Name, &task.Description, &task.Report, &task.Assigner, &task.Assignee, &task.Review, &task.ReviewAt, &task.Comment, &task.Proof, &task.StartAt, &task.OpenAt, &task.OpenAt, &task.OpenFrom, &task.Status, &task.IsClosed)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		tasks = append(tasks, task)

	}

	json.NewEncoder(w).Encode(tasks)
}

func createTask(w http.ResponseWriter, r *http.Request, user User) {
	var newTask Task
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	json.Unmarshal(reqBody, &newTask)

	if newTask.Name == nil && newTask.Description == nil {
		responseMessage(w, http.StatusBadRequest, "Task's name and task's description must not be empty!")
		return
	}

	var stt int = 0
	newTask.Status = &stt

	if !*user.IsAdmin {
		if user.GroupId != nil {
			group, err := getOneGroup(*user.GroupId)

			if err != nil {
				responseInternalError(w, err)
				return
			}

			if *group.ManagerId != *user.Id {
				newTask.Assignee = user.Id
				stt = 0
				goto insert
			}
		}
	}

	stt = 1
	newTask.Assigner = user.Id
	if newTask.Assignee == nil {
		responseMessage(w, http.StatusBadRequest, "Assignee must not be empty!")
		return
	}

	if newTask.StartAt == nil {
		responseMessage(w, http.StatusBadRequest, "Start must not be empty!")
		return
	}

	if newTask.StopAt == nil {
		responseMessage(w, http.StatusBadRequest, "Stop must not be empty!")
		return
	}

	if *newTask.StopAt <= *newTask.StartAt {
		responseMessage(w, http.StatusBadRequest, "Stop must happen after start!")
		return
	}

insert:

	// Check open_from
	// Check assignee

	_, err = db.Query("INSERT INTO `Tasks`(`name`, `description`,`assignee`,`status`,`assigner`,`open_at`,`open_from`) VALUES(?,?,?,?,?,?,?)", newTask.Name, newTask.Description, newTask.Assignee, newTask.Status, newTask.Assigner, time.Now().Unix(), newTask.OpenFrom)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(MessageResponse{
		Message: "New Task created successfully!",
	})
}

func routerGetOneTask(w http.ResponseWriter, r *http.Request, user User) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Id must be an integer!")
		return
	}

	task, err := getOneTask(id)
	if err != nil {
		response(w, http.StatusNotFound, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

func updateTask(w http.ResponseWriter, r *http.Request, user User) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Id must be an integer!")
		return
	}

	var taskToUpdate Task

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	json.Unmarshal(reqBody, &taskToUpdate)

	if taskToUpdate.Name == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Task's name must not be empty!",
		})
		return
	}

	thisTask, err := getOneTask(id)
	if err != nil {
		responseMessage(w, http.StatusNotFound, "Invalid task id")
		return
	}

	if thisTask.Assigner != nil && *thisTask.Assigner == *user.Id {
		_, err = db.Query("UPDATE `tasks` SET `name` = ?, `description` = ?, `report` = ? WHERE `id` = ?", taskToUpdate.Name, taskToUpdate.Description, taskToUpdate.Report, id)
	} else {
		_, err = db.Query("UPDATE `tasks` SET `report` = ? WHERE `id` = ?", taskToUpdate.Report, id)
	}
	if err != nil {
		responseInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(MessageResponse{
		Message: "Task updated!",
	})
}

func deleteTask(w http.ResponseWriter, r *http.Request, user User) {
	idGr := mux.Vars(r)["id"]

	_, err := db.Query("DELETE FROM `tasks` WHERE `id` = ?", idGr)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(MessageResponse{
		Message: "Task deleted!",
	})
}
