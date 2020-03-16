package app

import (
	"database/sql"
	"io/ioutil"
	"log"
	"net/http"
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

func getAllTasks(w http.ResponseWriter, r *http.Request, user User) {

	db, dbClose := openConnection()
	if db == nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		return
	}
	defer dbClose()

	results, err := db.Query("SELECT `id`, `name`, `description`,`report`,`assigner`,`assignee`,`review`,`review_at`,`comment`,`confirmation`,`start_at`,`close_at`,`open_at`,`open_from`,`status`,`is_closed` FROM `tasks`")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		log.Printf(err.Error())
		return
	}

	var tasks []Task

	for results.Next() {
		var task Task

		err = results.Scan(&task.Id, &task.Name, &task.Description, &task.Report, &task.Assigner, &task.Assignee, &task.Review, &task.ReviewAt, &task.Comment, &task.Confirmation, &task.StartAt, &task.OpenAt, &task.OpenAt, &task.OpenFrom, &task.Status, &task.IsClosed)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(MessageResponse{
				Message: "Internal error!",
			})
			log.Printf(err.Error())
			return
		}

		tasks = append(tasks, task)

	}

	json.NewEncoder(w).Encode(tasks)
}

func listCanBeOpenedFrom(w http.ResponseWriter, r *http.Request, user User) {
	db, dbClose, err := openConnection()
	if err != nil {
		responseInternalError(w, err)
		return
	}
	defer dbClose()

	query := "SELECT `id`, `name` FROM `tasks` WHERE (`is_closed` = TRUE AND `status` != 4) OR `status` != 5"

	if isManager(*user.Id) {
		results, err = db.Query(query+" `assigner` = ?", user.Id)
	} else {

		results, err = db.Query(query+" `assigner` = ?", user.Id)
	}
}

func listAssignableUsers(w http.ResponseWriter, r *http.Request, user User) {
	db, dbClose, err := openConnection()
	if err != nil {
		responseInternalError(w, err)
		return
	}
	defer dbClose()

	// fmt.Printf("%+v\n", *user.GroupId)

	var results *sql.Rows

	if *user.IsAdmin == true {
		results, err = db.Query("SELECT `id`, `full_name`, `username`, `group_id`, `role` FROM `users`")
	} else {
		results, err = db.Query("SELECT `id`, `full_name`, `username`, `group_id`, `role` FROM `users` WHERE `group_id` = ?", user.GroupId)
	}
	if err != nil {
		responseInternalError(w, err)
		return
	}

	var users []User

	for results.Next() {
		var user User

		err = results.Scan(&user.Id, &user.FullName, &user.Username, &user.GroupId, &user.Role)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		users = append(users, user)

	}

	json.NewEncoder(w).Encode(users)
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
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Task's name and task's description must not be empty!",
		})
		return
	}

	var status int = 1

	if *user.Role != 2 {
		newTask.Status = &status
		newTask.Assigner = user.Id
		if newTask.Assignee == nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(MessageResponse{
				Message: "Assignee must not be empty!",
			})
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

	_, err = db.Query("INSERT INTO `Tasks`(`name`, `description`,`assignee`,`status`,`assigner`) VALUES(?,?,?,?,?)", newTask.Name, newTask.Description, newTask.Assignee, newTask.Status)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(MessageResponse{
		Message: "New Task created successfully!",
	})
}

func getOneTask(w http.ResponseWriter, r *http.Request, user User) {
	idGr := mux.Vars(r)["id"]

	db, dbClose, err := openConnection()
	if err != nil {
		responseInternalError(w, err)
		return
	}
	defer dbClose()

	results, err := db.Query("SELECT `id`, `name`, `description`,`report`,`assigner`,`assignee`,`review`,`review_at`,`comment`,`confirmation`,`start_at`,`close_at`,`open_at`,`open_from`,`status`,`is_closed` FROM `tasks` WHERE `id` = ?", idGr)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	var task Task

	results.Next()

	err = results.Scan(&task.Id, &task.Name, &task.Description, &task.Report, &task.Assigner, &task.Assignee, &task.Review, &task.ReviewAt, &task.Comment, &task.Confirmation, &task.StartAt, &task.OpenAt, &task.OpenAt, &task.OpenFrom, &task.Status, &task.IsClosed)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

func updateTask(w http.ResponseWriter, r *http.Request, user User) {
	idGr := mux.Vars(r)["id"]

	var task Task

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseInternalError(w, err)
		return
	}

	json.Unmarshal(reqBody, &task)

	if task.Name == nil {
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

	if *user.Role == 2 {
		_, err = db.Query("UPDATE `tasks` SET `name` = ?, `description` = ?, `report` = ? WHERE `id` = ?", task.Name, task.Description, task.Report, idGr)
	} else {
		_, err = db.Query("UPDATE `tasks` SET `report` = ? WHERE `id` = ?", task.Report, idGr)
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
