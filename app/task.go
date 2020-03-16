package app

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Task struct {
	Id           *string `json:"id,omitempty"`
	Name         *string `json:"name"`
	Description  *string `json:"description"`
	Report       *string `json:"report"`
	Assigner     *string `json:"assigner"`
	Assignee     *string `json:"assignee"`
	Review       *string `json:"review"`
	ReviewAt     *string `json:"review_at"`
	Comment      *string `json:"comment"`
	Confirmation *string `json:"confirmation"`
	StartAt      *string `json:"start_at"`
	CloseAt      *string `json:"close_at"`
	OpenAt       *string `json:"open_at"`
	OpenFrom     *string `json:"open_from"`
	Status       *string `json:"status"`
	IsClosed     *bool   `json:"is_closed"`
}

func getAllTasks(w http.ResponseWriter, r *http.Request, id int) {

	db, dbClose := openConnection()
	if db == nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		return
	}
	defer dbClose()

	results, err := db.Query("SELECT `id`, `name`, `description` FROM `Tasks`")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		log.Printf(err.Error())
		return
	}

	var Tasks []Task

	for results.Next() {
		var Task Task

		err = results.Scan(&Task.Id, &Task.Name, &Task.Description)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(MessageResponse{
				Message: "Internal error!",
			})
			log.Printf(err.Error())
			return
		}

		Tasks = append(Tasks, Task)

	}

	json.NewEncoder(w).Encode(Tasks)
}

func createTask(w http.ResponseWriter, r *http.Request, id int) {
	var newTask Task
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		log.Printf(err.Error())
		return
	}

	json.Unmarshal(reqBody, &newTask)

	if newTask.Name == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Task's name must not be empty!",
		})
		return
	}

	db, dbClose := openConnection()
	if db == nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		return
	}
	defer dbClose()

	_, err = db.Query("INSERT INTO `Tasks`(`name`, `description`) VALUES(?,?)", newTask.Name, newTask.Description)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		log.Printf(err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(MessageResponse{
		Message: "New Task created successfully!",
	})
}

func getOneTask(w http.ResponseWriter, r *http.Request, id int) {
	idGr := mux.Vars(r)["id"]

	db, dbClose := openConnection()
	if db == nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		return
	}
	defer dbClose()

	results, err := db.Query("SELECT `id`,`name`,`description` FROM `Tasks` WHERE `id` = ?", idGr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		log.Printf(err.Error())
		return
	}

	var Task Task

	results.Next()

	err = results.Scan(&Task.Id, &Task.Name, &Task.Description)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		log.Printf(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Task)
}

func updateTask(w http.ResponseWriter, r *http.Request, id int) {
	idGr := mux.Vars(r)["id"]

	var Task Task

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		log.Printf(err.Error())
		return
	}

	json.Unmarshal(reqBody, &Task)

	if Task.Name == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Task's name must not be empty!",
		})
		return
	}

	db, dbClose := openConnection()
	if db == nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		return
	}
	defer dbClose()

	_, err = db.Query("UPDATE `Tasks` SET `name` = ?, `description` = ? WHERE `id` = ?", Task.Name, Task.Description, idGr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		log.Printf(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(MessageResponse{
		Message: "Task updated!",
	})
}

func deleteTask(w http.ResponseWriter, r *http.Request, id int) {
	idGr := mux.Vars(r)["id"]

	db, dbClose := openConnection()
	if db == nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		return
	}
	defer dbClose()

	_, err := db.Query("DELETE FROM `Tasks` WHERE `id` = ?", idGr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "Internal error!",
		})
		log.Printf(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(MessageResponse{
		Message: "Task deleted!",
	})
}
