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
	var task Task

	results := db.QueryRow("SELECT `id`, `name`, `description`,`report`,`assigner`,`assignee`,`review`,`comment`,`proof`,`start_at`,`stop_at`,`close_at`,`open_at`,`open_from`,`status`,`is_closed` FROM `tasks` WHERE `id` = ?", id)
	err := results.Scan(&task.Id, &task.Name, &task.Description, &task.Report, &task.Assigner, &task.Assignee, &task.Review, &task.Comment, &task.Proof, &task.StartAt, &task.StopAt, &task.CloseAt, &task.OpenAt, &task.OpenFrom, &task.Status, &task.IsClosed)
	if err == sql.ErrNoRows {
		return nil, errors.New("Invalid task id")
	} else if err != nil {
		return nil, err
	}

	return &task, nil
}

// -----------------------------------------------------------------------------

func getReopenableTasks(w http.ResponseWriter, r *http.Request, user User) {

	query := "SELECT `id`, `name` FROM `tasks` WHERE (`is_closed` = TRUE AND `status` != 4) OR `status` != 5"

	var results *sql.Rows
	var err error
	var stuffs []interface{}

	if !*user.IsAdmin {

		if user.GroupId != nil {
			var group *Group
			group, err = getOneGroup(*user.GroupId)
			if err != nil {
				responseInternalError(w, err)
				return
			}

			query = query + " AND `assigner` = ?"
			stuffs = append(stuffs, *group.ManagerId)

		} else {
			responseMessage(w, http.StatusForbidden, "Please join a group to create tasks")
			return
		}
	}

	results, err = db.Query(query, stuffs...)
	if err != nil {
		responseInternalError(w, err)
		return
	}
	defer results.Close()

	tasks := make([]Task, 0)

	for results.Next() {
		var task Task

		err := results.Scan(&task.Id, &task.Name)
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

	query := "SELECT `id`, `full_name`, `username`, `group_id` FROM `users` WHERE `is_admin` = 0"

	if *user.IsAdmin {
		results, err = db.Query(query+" AND `id` != ?", user.Id)
		if err != nil {
			responseInternalError(w, err)
			return
		}
		defer results.Close()
	} else if user.GroupId != nil {
		group, err := getOneGroup(*user.GroupId)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		if *group.ManagerId == *user.Id {
			results, err = db.Query(query+" AND `group_id` = ?", user.GroupId)
			if err != nil {
				responseInternalError(w, err)
				return
			}
			defer results.Close()
		} else {
			results, err = db.Query(query+" AND `id` = ?", user.Id)
			if err != nil {
				responseInternalError(w, err)
				return
			}
			defer results.Close()
		}
	} else {
		responseMessage(w, http.StatusForbidden, "Please join a group to create tasks")
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

func checkTask(w http.ResponseWriter, r *http.Request, user User) { // manager
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

	_, err = db.Exec("UPDATE `tasks` SET `"+attr+"` = 1 WHERE `id` = ?", id)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	responseMessage(w, http.StatusOK, "Task checked!")
}

func getPermission(w http.ResponseWriter, r *http.Request, user User) { // user
	if *user.IsAdmin {
		responseMessage(w, http.StatusOK, "manage")
		return
	}

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

	if task.Assigner != nil && *task.Assigner == *user.Id {
		responseMessage(w, http.StatusOK, "manage")
		return
	}

	if *task.Assignee == *user.Id {
		responseMessage(w, http.StatusOK, "do")
		return
	}

	responseMessage(w, http.StatusNotFound, "nope")
}

func startTask(w http.ResponseWriter, r *http.Request, user User) { // user
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

	if *task.Status != 1 || *task.IsClosed || (*task.Assignee != *user.Id && *task.Assigner != *user.Id && !*user.IsAdmin) {
		responseMessage(w, http.StatusBadRequest, "Cannot start this task!")
		return
	}

	_, err = db.Exec("UPDATE `tasks` SET `status` = 2, `start_at` = ? WHERE `id` = ?", time.Now().Unix(), id)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	responseMessage(w, http.StatusOK, "Task checked!")
}

func closeTask(w http.ResponseWriter, r *http.Request, user User) { // manager
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

	if *task.IsClosed || (!*user.IsAdmin && *user.Id != *task.Assigner) {
		responseMessage(w, http.StatusBadRequest, "Cannot close this task!")
		return
	}

	_, err = db.Exec("UPDATE `tasks` SET `is_closed` = 1, `close_at` = ? WHERE `id` = ?", time.Now().Unix(), id)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	responseMessage(w, http.StatusOK, "Task closed!")
}

func confirmTask(w http.ResponseWriter, r *http.Request, user User) { // user
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

	if *task.Status != 2 || *task.IsClosed || (*task.Assignee != *user.Id && *task.Assigner != *user.Id && !*user.IsAdmin) {
		logg("WTFFFFFF")
		responseMessage(w, http.StatusBadRequest, "Cannot confirm this task!")
		return
	}

	r.ParseMultipartForm(10 << 20)

	file, _, err := r.FormFile("proof")
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Please upload proof!")
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
	_, err = db.Exec("UPDATE `tasks` SET `status` = ?, `proof` = ? WHERE `id` = ?", stt, filename, id)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	responseMessage(w, http.StatusOK, "Confirm task successfully!")
}

func verifyTask(w http.ResponseWriter, r *http.Request, user User) { // manager
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
		_, err = db.Exec("UPDATE `tasks` SET `status` = 2 WHERE `id` = ?", id)
	} else {
		_, err = db.Exec("UPDATE `tasks` SET `is_closed` = TRUE, `close_at` = ?  WHERE`id` = ?", time.Now().Unix(), id)
	}

	if err != nil {
		responseInternalError(w, err)
		return
	}

	responseMessage(w, http.StatusOK, "Verify task successfully")
}

// -----------------------------------------------------------------------------

func getAllTasks(w http.ResponseWriter, r *http.Request, user User) {

	var rassignee []interface{}
	var rassigner []interface{}
	var rstt []interface{}
	var rcls []interface{}
	var endDate interface{}
	var startDate interface{}

	if r.URL.Query().Get("assignee") != "" {
		assignee := r.URL.Query().Get("assignee")
		splited := strings.Split(assignee, ",")

		for _, s := range splited {
			r, err := strconv.Atoi(s)

			if err != nil {
				responseMessage(w, http.StatusBadRequest, "Assignees must be integers")
				return
			}

			rassignee = append(rassignee, r)
		}
	}

	if r.URL.Query().Get("assigner") != "" {
		assigner := r.URL.Query().Get("assigner")
		splited := strings.Split(assigner, ",")

		for _, s := range splited {
			r, err := strconv.Atoi(s)

			if err != nil {
				responseMessage(w, http.StatusBadRequest, "Assigners must be integers")
				return
			}

			rassigner = append(rassigner, r)
		}
	}

	if r.URL.Query().Get("status") != "" {
		statuses := r.URL.Query().Get("status")
		splited := strings.Split(statuses, ",")

		var tuses []int

		for _, s := range splited {
			r, err := strconv.Atoi(s)
			if err != nil {
				responseMessage(w, http.StatusBadRequest, "Status must be integers")
				return
			}
			tuses = append(tuses, r)
		}

		var stt []int
		var cls []int

		for _, s := range tuses {
			switch s {
			case -1: // 0, 1
				stt = append(stt, 0)
				cls = append(cls, 1)
			case 0: // 0, 0
				stt = append(stt, 0)
				cls = append(cls, 0)
			case 1: // 1, 0
				stt = append(stt, 1)
				cls = append(cls, 0)
			case 2: // 2, 0
				stt = append(stt, 2)
				cls = append(cls, 0)
			case 3: // 3|4, 0
				stt = append(stt, 3, 4)
				cls = append(cls, 0)
			case 4: // 3, 1
				stt = append(stt, 3)
				cls = append(cls, 1)
			case 5: // 4, 1
				stt = append(stt, 4)
				cls = append(cls, 1)
			case 6: // 5
				stt = append(stt, 5)
			case 7: // 1|2, 1
				stt = append(stt, 1, 2)
				cls = append(cls, 1)
			}
		}

		for _, s := range unique(stt) {
			rstt = append(rstt, s)
		}

		for _, s := range unique(cls) {
			rcls = append(rcls, s)
		}
	}

	if r.URL.Query().Get("deadline") != "" {
		deadline := r.URL.Query().Get("deadline")
		splited := strings.Split(deadline, ",")

		if len(splited) < 2 {
			responseMessage(w, http.StatusBadRequest, "Bad deadline")
			return
		}

		parse_time, err := strconv.Atoi(splited[0])
		if err != nil {
			responseMessage(w, http.StatusBadRequest, "Bad deadline")
			return
		}
		startDate = parse_time

		parse_time, err = strconv.Atoi(splited[1])
		if err != nil {
			responseMessage(w, http.StatusBadRequest, "Bad deadline")
			return
		}
		endDate = parse_time
	}

	// logg(rassignee)
	// logg(rassigner)
	// logg(rstt)
	// logg(rcls)
	// logg(startDate)
	// logg(endDate)

	query := "SELECT `id`, `name`, `description`,`report`,`assigner`,`assignee`,`review`,`comment`,`proof`,`start_at`,`stop_at`,`close_at`,`open_at`,`open_from`,`status`,`is_closed` FROM `tasks`"

	var stuffs []interface{}
	var and bool
	if len(rstt) > 0 || len(rcls) > 0 || len(rassignee) > 0 || len(rassigner) > 0 || (startDate != nil && endDate != nil) {
		query = query + " WHERE"
		if len(rstt) > 0 {
			query = query + " `status` IN (?" + strings.Repeat(",?", len(rstt)-1) + ")"
			stuffs = append(stuffs, rstt...)
			and = true
		}

		if len(rcls) > 0 {
			if and {
				query = query + " AND"
			} else {
				and = true
			}
			query = query + " `is_closed` IN (?" + strings.Repeat(",?", len(rcls)-1) + ")"
			stuffs = append(stuffs, rcls...)
		}

		if len(rassignee) > 0 {
			if and {
				query = query + " AND"
			} else {
				and = true
			}
			query = query + " `assignee` IN (?" + strings.Repeat(",?", len(rassignee)-1) + ")"
			stuffs = append(stuffs, rassignee...)
		}

		if len(rassigner) > 0 {
			if and {
				query = query + " AND"
			} else {
				and = true
			}
			query = query + " `assigner` IN (?" + strings.Repeat(",?", len(rassigner)-1) + ")"
			stuffs = append(stuffs, rassigner...)
		}

		if startDate != nil && endDate != nil {
			if and {
				query = query + " AND"
			}
			query = query + " `stop_at` >= ? AND `stop_at` <= ?"
			stuffs = append(stuffs, startDate, endDate)
		}
	}

	query = query + " ORDER BY stop_at ASC"

	logg(query)

	results, err := db.Query(query, stuffs...)
	if err != nil {
		responseInternalError(w, err)
		return
	}
	defer results.Close()

	tasks := make([]Task, 0)

	for results.Next() {
		var task Task

		err = results.Scan(&task.Id, &task.Name, &task.Description, &task.Report, &task.Assigner, &task.Assignee, &task.Review, &task.Comment, &task.Proof, &task.StartAt, &task.StopAt, &task.CloseAt, &task.OpenAt, &task.OpenFrom, &task.Status, &task.IsClosed)
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

	logg(newTask)

	if newTask.Name == nil && newTask.Description == nil {
		responseMessage(w, http.StatusBadRequest, "Task's name and task's description must not be empty!")
		return
	}

	var stt int = 0
	newTask.Status = &stt

	if !*user.IsAdmin {
		if user.GroupId != nil {
			logg(*user.GroupId)
			group, err := getOneGroup(*user.GroupId)

			if err != nil {
				responseInternalError(w, err)
				return
			}

			if *group.ManagerId != *user.Id { // not manager
				newTask.Assignee = user.Id
				stt = 0
				goto insert
			}
		} else {
			responseMessage(w, http.StatusForbidden, "Please join a group to create tasks")
			return
		}
	}

	// Admin or manager
	stt = 1
	newTask.Assigner = user.Id
	if newTask.Assignee == nil {
		responseMessage(w, http.StatusBadRequest, "Assignee must not be empty!")
		return
	}

	if newTask.StopAt == nil {
		responseMessage(w, http.StatusBadRequest, "Stop must not be empty!")
		return
	}

	if *newTask.StopAt <= time.Now().Unix() {
		responseMessage(w, http.StatusBadRequest, "Stop must happen in the future!")
		return
	}

insert:

	// Check open_from
	// Check assignee

	results, err := db.Exec("INSERT INTO `Tasks`(`name`, `description`,`assignee`,`status`,`assigner`,`open_at`,`open_from`) VALUES(?,?,?,?,?,?,?)", newTask.Name, newTask.Description, newTask.Assignee, newTask.Status, newTask.Assigner, time.Now().Unix(), newTask.OpenFrom)
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
		_, err = db.Exec("UPDATE `tasks` SET `name` = ?, `description` = ?, `report` = ? WHERE `id` = ?", taskToUpdate.Name, taskToUpdate.Description, taskToUpdate.Report, id)
	} else if thisTask.Assignee != nil && *thisTask.Assignee == *user.Id {
		_, err = db.Exec("UPDATE `tasks` SET `report` = ? WHERE `id` = ?", taskToUpdate.Report, id)
	} else {
		responseMessage(w, http.StatusUnauthorized, "Cannot edit")
		return
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
	if *user.IsAdmin {
		idGr := mux.Vars(r)["id"]

		_, err := db.Exec("DELETE FROM `tasks` WHERE `id` = ?", idGr)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Task deleted!",
		})
	}
}
