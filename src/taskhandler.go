package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type taskModel struct {
	Taskid      uuid.UUID `json:"taskId"`
	Description string    `json:"description"`
	CreatedAt   string    `json:"createdAt"`
	UpdatedAt   string    `json:"updatedAt"`
	ValidTill   *string   `json:"validTill"`
	Userid      uuid.UUID `json:"userId"`
}

func fetchTasks(w http.ResponseWriter, conn *sql.DB, r *http.Request) {
	//userid := "b2a0b669-ebf4-433f-8f59-898a79f61d4e"
	userid := r.Context().Value(userContext).(string)
	data, err := conn.Query("SELECT * FROM users.tasks WHERE userid=$1 and validtill is null", userid)
	if err != nil {
		responseWithJson(w, http.StatusBadRequest, err)
		log.Printf("Cant fetch data. Error: %v", err)
		return
	}
	defer data.Close()

	tasks := make([]taskModel, 0)
	for data.Next() {
		task := taskModel{}
		err = data.Scan(&task.Taskid, &task.Description, &task.CreatedAt, &task.UpdatedAt, &task.ValidTill, &task.Userid)
		if err != nil {
			log.Println(err)
			responseWithJson(w, http.StatusBadRequest, err)
			return
		}
		tasks = append(tasks, task)
	}

	if len(tasks) == 0 {
		responseWithJson(w, http.StatusOK, map[string]string{
			"body": "No records found",
		})
	} else {
		responseWithJson(w, http.StatusOK, tasks)
	}
}

func addTasks(w http.ResponseWriter, conn *sql.DB, r *http.Request) {
	//userid := r.Context().Value(userContextKey).(string)
	var payload taskModel
	err := parsePayload(r.Body, &payload)
	if err != nil {
		responseWithJson(w, http.StatusBadRequest, err)
		log.Printf("Error while parsing payload. Error: %v", err)
		return
	}

	sqlQuery := `insert into users.tasks values ($1, $2, $3, $4, $5, $6) returning taskid`
	id := ""
	generateTaskId := uuid.New()
	err = conn.QueryRow(sqlQuery, generateTaskId, payload.Description, time.Now(), time.Now(), nil, payload.Userid).Scan(&id)
	if err != nil {
		responseWithJson(w, http.StatusBadRequest, err)
		log.Printf("Error while inserting record. Error: %v", err)
		return
	}

	//fetch the newly added record to send the response
	sqlQuery = `select * from users.tasks where taskid=$1`
	data, err := conn.Query(sqlQuery, generateTaskId)
	if err != nil {
		responseWithJson(w, http.StatusBadRequest, err)
		log.Printf("Cant fetch data. Error: %v", err)
		return
	}
	defer data.Close()

	task := taskModel{}
	for data.Next() {
		err := data.Scan(&task.Taskid, &task.Description, &task.CreatedAt, &task.UpdatedAt, &task.ValidTill, &task.Userid)
		if err != nil {
			log.Println(err)
			responseWithJson(w, http.StatusBadRequest, err)
			return
		}
	}

	responseWithJson(w, http.StatusCreated, task)
}

func updateTasksWithId(w http.ResponseWriter, conn *sql.DB, r *http.Request) {
	userid := r.Context().Value(userContext).(string)
	var payload taskModel
	err := parsePayload(r.Body, &payload)
	if err != nil {
		responseWithJson(w, http.StatusBadRequest, err)
		log.Printf("Error while parsing payload. Error: %v", err)
		return
	}

	sqlQuery := `update users.tasks set description=$1, updatedat=$2 where userid=$3 and taskid=$4;`
	_, err = conn.Exec(sqlQuery, payload.Description, time.Now(), userid, payload.Taskid)
	if err != nil {
		responseWithJson(w, http.StatusBadRequest, err)
		log.Printf("Error while inserting record. Error: %v", err)
		return
	}

	//fetch the newly update record to send the response
	sqlQuery = `select * from users.tasks where userid=$1 and taskid=$2`
	data, err := conn.Query(sqlQuery, userid, payload.Taskid)
	if err != nil {
		responseWithJson(w, http.StatusBadRequest, err)
		log.Printf("Cant fetch data. Error: %v", err)
		return
	}
	defer data.Close()

	task := taskModel{}
	for data.Next() {
		err := data.Scan(&task.Taskid, &task.Description, &task.CreatedAt, &task.UpdatedAt, &task.ValidTill, &task.Userid)
		if err != nil {
			log.Println(err)
			responseWithJson(w, http.StatusBadRequest, err)
			return
		}
	}

	responseWithJson(w, http.StatusCreated, task)
}

func deleteTasksWithId(w http.ResponseWriter, conn *sql.DB, r *http.Request) {
	userid := r.Context().Value(userContext).(string)
	var payload taskModel
	err := parsePayload(r.Body, &payload)
	if err != nil {
		responseWithJson(w, http.StatusBadRequest, err)
		log.Printf("Error while parsing payload. Error: %v", err)
		return
	}

	sqlQuery := `update users.tasks set validtill=$1 where userid=$2 and taskid=$3;`
	_, err = conn.Exec(sqlQuery, time.Now(), userid, payload.Taskid)
	if err != nil {
		responseWithJson(w, http.StatusBadRequest, err)
		log.Printf("Error while inserting record. Error: %v", err)
		return
	}

	//responseWithJson(w, http.StatusCreated, payload)
	w.WriteHeader(http.StatusCreated)
}

func getTasksWithId(w http.ResponseWriter, conn *sql.DB, r *http.Request) {
	taskidFromURL := chi.URLParam(r, "id")
	//userid := r.Context().Value(userContextKey).(string)
	data, err := conn.Query("SELECT * FROM users.tasks WHERE taskid=$1 and validtill is null", taskidFromURL)
	if err != nil {
		responseWithJson(w, http.StatusBadRequest, err)
		log.Printf("Cant fetch data. Error: %v", err)
		return
	}
	defer data.Close()

	//tasks := make([]taskModel, 0)
	task := taskModel{}
	for data.Next() {
		err = data.Scan(&task.Taskid, &task.Description, &task.CreatedAt, &task.UpdatedAt, &task.ValidTill, &task.Userid)
		if err != nil {
			log.Println(err)
			responseWithJson(w, http.StatusBadRequest, err)
			return
		}
		//tasks = append(tasks, task)
	}
	if task.Description != "" {
		responseWithJson(w, http.StatusOK, task)
	} else {
		responseWithJson(w, http.StatusOK, map[string]string{
			"body": "No records found",
		})
	}
}
