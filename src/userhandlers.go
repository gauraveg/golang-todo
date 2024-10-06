package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type userModel struct {
	Userid    uuid.UUID `json:"userid"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt string    `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
}

func NullString(s *string) sql.NullString {
	sen := *s
	if len(sen) == 0 {
		return sql.NullString{}
	}

	return sql.NullString{
		String: sen,
		Valid:  true,
	}
}

func responseWithJson(w http.ResponseWriter, code int, payload interface{}) {
	w.WriteHeader(code)
	if payload != nil {
		err := json.NewEncoder(w).Encode(payload)
		if err != nil {
			log.Printf("Cannot parse payload : %v", err)
			return
		}
	}
}

func parsePayload(body io.Reader, out interface{}) error {
	err := json.NewDecoder(body).Decode(out)
	if err != nil {
		return err
	}

	return nil
}

func fetchUsers(w http.ResponseWriter, conn *sql.DB) {
	data, err := conn.Query("select * from users.users")
	if err != nil {
		responseWithJson(w, http.StatusBadRequest, err)
		log.Printf("Cant fetch data. Error: %v", err)
		return
	}
	defer data.Close()

	users := make([]userModel, 0)
	for data.Next() {
		user := userModel{}
		err := data.Scan(&user.Userid, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Println(err)
			responseWithJson(w, http.StatusBadRequest, err)
			return
		}

		users = append(users, user)
	}

	responseWithJson(w, http.StatusOK, users)
}

func addUser(w http.ResponseWriter, conn *sql.DB, r *http.Request) {
	var payload userModel
	err := parsePayload(r.Body, &payload)
	if err != nil {
		responseWithJson(w, http.StatusBadRequest, err)
		log.Printf("Error while parsing payload. Error: %v", err)
		return
	}

	sqlQuery := `insert into users.users values ($1, $2, $3, $4, $5, $6) returning userid`
	id := ""
	err = conn.QueryRow(sqlQuery, uuid.New(), payload.Name, payload.Email, payload.Password, time.Now(), time.Now()).Scan(&id)
	if err != nil {
		responseWithJson(w, http.StatusBadRequest, err)
		log.Printf("Error while inserting record. Error: %v", err)
		return
	}

	responseWithJson(w, http.StatusCreated, payload)
}
