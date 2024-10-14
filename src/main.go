package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var conn *sql.DB

const userContext string = "userContext"

type loginResponse struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type sessionToken struct {
	Status string `json:"status"`
	Token  string `json:"token"`
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionToken := r.Header.Get("token")
		if sessionToken == "" {
			// responseWithJson(w, http.StatusUnauthorized, map[string]string{
			// 	"status": "login failed",
			// })
			log.Fatal("Login failed. Session token not present")
		}
		userId, err := getUserIdForToken(sessionToken)
		if err != nil {
			responseWithJson(w, http.StatusUnauthorized, err)
			log.Printf("Cannot fetch userId from the session token. Error: %v", err)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), userContext, userId))
		next.ServeHTTP(w, r)
	})
}

func getUserIdForToken(sessionToken string) (string, error) {
	sqlQuery := `select userid from users.session where sessionid=$1 and validtill is null`
	var userid string
	err := conn.QueryRow(sqlQuery, sessionToken).Scan(&userid)
	return userid, err
}

func logout(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.Header.Get("token")
	sqlQuery := `update users.session set validtill=$1 where sessionid=$2`
	_, err := conn.Exec(sqlQuery, time.Now(), sessionToken)
	if err != nil {
		responseWithJson(w, http.StatusBadRequest, err)
		log.Printf("Error while updating record. Error: %v", err)
		return
	}
	responseWithJson(w, http.StatusCreated, map[string]string{
		"status": "Logout success",
	})
}

// Loginhandlers
func login(w http.ResponseWriter, r *http.Request) {
	var payload loginResponse
	err := parsePayload(r.Body, &payload)
	if err != nil {
		responseWithJson(w, http.StatusBadRequest, err)
		log.Printf("Error while parsing payload. Error: %v", err)
		return
	}
	sqlQuery := `select userid from users.users where email=$1 and password=$2`
	userid := ""
	err = conn.QueryRow(sqlQuery, payload.Email, payload.Password).Scan(&userid)
	if err != nil {
		responseWithJson(w, http.StatusBadRequest, err)
		log.Printf("Error while inserting record. Error: %v", err)
		return
	}

	sqlQueryINS := `insert into users.session values ($1, $2);`
	sessionId := uuid.New()
	_, err = conn.Exec(sqlQueryINS, sessionId, userid)
	if err != nil {
		responseWithJson(w, http.StatusBadRequest, err)
		log.Printf("Error while inserting record. Error: %v", err)
		return
	}

	responseWithJson(w, http.StatusCreated, sessionToken{
		Status: "Login success",
		Token:  sessionId.String(),
	})
}

// status handler
func handlerStatus(w http.ResponseWriter, r *http.Request) {
	responseWithJson(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// User handlers
func handlerGETUser(w http.ResponseWriter, r *http.Request) {
	fetchUsers(w, conn)
}

func handlerPOSTUser(w http.ResponseWriter, r *http.Request) {
	addUser(w, conn, r)
}

func handlerUser(w http.ResponseWriter, r *http.Request) {
	responseWithJson(w, http.StatusOK, map[string]string{
		"user list -> GET": "/users",
		"Add user -> POST": "/users/add",
	})
}

// Task handlers
func handlerTask(w http.ResponseWriter, r *http.Request) {
	responseWithJson(w, http.StatusOK, map[string]string{
		"task list -> GET":      "/tasks",
		"Add task -> POST":      "/tasks/add",
		"Update task -> PUT":    "/tasks/{id}/update",
		"Delete task -> DELETE": "/tasks/{id}/",
	})
}

func handlerGETTask(w http.ResponseWriter, r *http.Request) {
	fetchTasks(w, conn, r)
}

func handlerPOSTTask(w http.ResponseWriter, r *http.Request) {
	addTasks(w, conn, r)
}

func handlerPUTTaskWithId(w http.ResponseWriter, r *http.Request) {
	updateTasksWithId(w, conn, r)
}

func handlerGETTaskWithId(w http.ResponseWriter, r *http.Request) {
	getTasksWithId(w, conn, r)
}

func handlerDELTaskWithId(w http.ResponseWriter, r *http.Request) {
	deleteTasksWithId(w, conn, r)
}

// Main func to call the router
func main() {
	var err error
	godotenv.Load()

	portVal := os.Getenv("PORT")
	dburl := os.Getenv("DB_URL")
	conn, err = sql.Open("pgx", dburl)
	if err != nil {
		log.Printf("Cannot connect database. Error: %v", err)
		return
	}
	defer conn.Close()
	log.Println("Database connected!")

	router := chi.NewRouter()
	router.Get("/status", handlerStatus)
	router.Post("/login", login)

	router.Route("/users", func(userRouter chi.Router) {
		userRouter.Get("/", handlerUser)
		userRouter.Get("/list", handlerGETUser)
		userRouter.Post("/add", handlerPOSTUser)
	})

	router.Route("/tasks", func(taskRouter chi.Router) {
		taskRouter.Use(AuthMiddleware)
		taskRouter.Get("/", handlerTask)
		taskRouter.Get("/list", handlerGETTask)
		taskRouter.Post("/add", handlerPOSTTask)
		taskRouter.Route("/{id}", func(taskIdRouter chi.Router) {
			taskIdRouter.Get("/", handlerGETTaskWithId)
			taskIdRouter.Put("/", handlerPUTTaskWithId)
			taskIdRouter.Delete("/", handlerDELTaskWithId)
		})
	})

	router.Post("/logout", logout)

	log.Printf("Server has started on port: %v", portVal)
	http.ListenAndServe(":"+portVal, router)
}
