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
	"golang.org/x/crypto/bcrypt"
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

	//Password hashing
	hashedPwd := passwordhashing(payload.Password)

	sqlQuery := `insert into users.users values ($1, $2, $3, $4, $5, $6) returning userid`
	id := ""
	generateUserId := uuid.New()
	err = conn.QueryRow(sqlQuery, generateUserId, payload.Name, payload.Email, hashedPwd, time.Now(), time.Now()).Scan(&id)
	if err != nil {
		responseWithJson(w, http.StatusBadRequest, err)
		log.Printf("Error while inserting record. Error: %v", err)
		return
	}

	//fetch the newly added record to send the response
	sqlQuery = `select * from users.users where userid=$1`
	data, err := conn.Query(sqlQuery, generateUserId)
	if err != nil {
		responseWithJson(w, http.StatusBadRequest, err)
		log.Printf("Cant fetch data. Error: %v", err)
		return
	}
	defer data.Close()
	//users := make([]userModel, 0)
	user := userModel{}
	for data.Next() {
		//user := userModel{}
		err := data.Scan(&user.Userid, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Println(err)
			responseWithJson(w, http.StatusBadRequest, err)
			return
		}
		//users = append(users, user)
	}

	responseWithJson(w, http.StatusCreated, user)
}

func passwordhashing(pwd string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}

	return string(hash)
}

func verifyPwdHash(pwd string, userPwdhash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(userPwdhash), []byte(pwd))
	return err == nil
}
