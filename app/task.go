package app

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Task struct {
	Id           *int       `json:"id,omitempty"`
	Name         *string    `json:"name"`        // Updatable
	Description  *string    `json:"description"` // Updatanle
	Report       *string    `json:"report"`      // Updateable
	Assigner     *int       `json:"assigner"`
	Assignee     *int       `json:"assignee"`
	Review       *int       `json:"review"`
	ReviewAt     *time.Time `json:"review_at"`
	Comment      *string    `json:"comment"`
	Confirmation *string    `json:"confirmation"`
	StartAt      *time.Time `json:"start_at"`
	CloseAt      *time.Time `json:"close_at"`
	OpenAt       *time.Time `json:"open_at"`
	OpenFrom     *int       `json:"open_from"`
	Status       *int       `json:"status"`
	IsClosed     *bool      `json:"is_closed"`
}

func getOneTask(id int) (*Task, error) {
	db, dbClose, err := openConnection()
	if err != nil {
		return nil, err
	}
	defer dbClose()

	results, err := db.Query("SELECT `id`, `name`, `description`,`report`,`assigner`,`assignee`,`review`,`review_at`,`comment`,`confirmation`,`start_at`,`close_at`,`open_at`,`open_from`,`status`,`is_closed` FROM `tasks` WHERE `id` = ?", id)
	if err != nil {
		return nil, err
	}

	if results.Next() {
		var task Task

		err = results.Scan(&task.Id, &task.Name, &task.Description, &task.Report, &task.Assigner, &task.Assignee, &task.Review, &task.ReviewAt, &task.Comment, &task.Confirmation, &task.StartAt, &task.OpenAt, &task.OpenAt, &task.OpenFrom, &task.Status, &task.IsClosed)
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
	db, dbClose, err := openConnection()
	if err != nil {
		responseInternalError(w, err)
		return
	}
	defer dbClose()

	query := "SELECT `id`, `name` FROM `tasks` WHERE (`is_closed` = TRUE AND `status` != 4) OR `status` != 5"

	var mng int
	var results *sql.Rows

	if !*user.IsAdmin || user.GroupId != nil {
		mng, err = getManager(*user.GroupId)

		results, err = db.Query(query+" `assigner` = ?", mng)
		if err != nil {
			responseInternalError(w, err)
			return
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
	db, dbClose, err := openConnection()
	if err != nil {
		responseInternalError(w, err)
		return
	}
	defer dbClose()

	var results *sql.Rows

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

func getAllTasks(w http.ResponseWriter, r *http.Request, user User) {

	db, dbClose, err := openConnection()
	if err != nil {
		responseInternalError(w, err)
		return
	}
	defer dbClose()

	results, err := db.Query("SELECT `id`, `name`, `description`,`report`,`assigner`,`assignee`,`review`,`review_at`,`comment`,`confirmation`,`start_at`,`close_at`,`open_at`,`open_from`,`status`,`is_closed` FROM `tasks`")
	if err != nil {
		responseInternalError(w, err)
		return
	}

	tasks := make([]Task, 0)

	for results.Next() {
		var task Task

		err = results.Scan(&task.Id, &task.Name, &task.Description, &task.Report, &task.Assigner, &task.Assignee, &task.Review, &task.ReviewAt, &task.Comment, &task.Confirmation, &task.StartAt, &task.OpenAt, &task.OpenAt, &task.OpenFrom, &task.Status, &task.IsClosed)
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
		responseCustomError(w, http.StatusBadRequest, "Task's name and task's description must not be empty!")
		return
	}

	var status int = 1

	var mng int

	mng, err = getManager(*user.GroupId)

	isMng := *user.GroupId == mng

	if *user.IsAdmin || isMng {
		newTask.Status = &status
		newTask.Assigner = user.Id
		if newTask.Assignee == nil {
			responseCustomError(w, http.StatusBadRequest, "Assignee must not be empty!")
			return
		}
	} else {
		newTask.Assignee = user.Id
	}

	db, dbClose, err := openConnection()
	if err != nil {
		responseInternalError(w, err)
		return
	}
	defer dbClose()

	_, err = db.Query("INSERT INTO `Tasks`(`name`, `description`,`assignee`,`status`,`assigner`) VALUES(?,?,?,?,?)", newTask.Name, newTask.Description, newTask.Assignee, newTask.Status, newTask.Assigner)
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
		responseCustomError(w, http.StatusBadRequest, "Id must be an integer!")
		return
	}

	task, err := getOneTask(id)
	if err != nil {
		responseCustomError(w, http.StatusNotFound, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

func updateTask(w http.ResponseWriter, r *http.Request, user User) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		responseCustomError(w, http.StatusBadRequest, "Id must be an integer!")
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

	db, dbClose, err := openConnection()
	if err != nil {
		responseInternalError(w, err)
		return
	}
	defer dbClose()

	thisTask, err := getOneTask(id)
	if err != nil {
		responseCustomError(w, http.StatusNotFound, "Invalid task id")
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

	db, dbClose, err := openConnection()
	if err != nil {
		responseInternalError(w, err)
		return
	}
	defer dbClose()

	_, err = db.Query("DELETE FROM `tasks` WHERE `id` = ?", idGr)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(MessageResponse{
		Message: "Task deleted!",
	})
}
