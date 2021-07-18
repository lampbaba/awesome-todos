package app

import (
	"awesome/todos/model"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
)

var store = sessions.NewCookieStore([]byte("Security123!@#"))
var rd *render.Render = render.New()

type AppHandler struct {
	http.Handler
	db model.DBHandler
}

type Success struct {
	Success bool `json:"success"`
}

// 로그인 때문에 테스트가 되지 않으므로 함수를 variable로 선언한다.
var getSessionID = func(r *http.Request) string {
	session, err := store.Get(r, "session")
	if err != nil {
		return ""
	}

	id := session.Values["id"]
	if id == nil {
		return ""
	}
	// picture := fmt.Sprintf("%v", session.Values["picture"])
	// picture := session.Values["picture"]

	// return id.(string), picture.(string)
	log.Println("sessionId: " + id.(string))
	return id.(string)
}

func (appHandler *AppHandler) indexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/todo.html", http.StatusTemporaryRedirect)
}

func (appHandler *AppHandler) getTodoListHandler(w http.ResponseWriter, r *http.Request) {
	sessionId := getSessionID(r)
	list := appHandler.db.GetTodos(sessionId)
	rd.JSON(w, http.StatusOK, list)
}

func (appHandler *AppHandler) addTodoHandler(w http.ResponseWriter, r *http.Request) {
	sessionId := getSessionID(r)
	name := r.FormValue("name")
	todo := appHandler.db.AddTodo(sessionId, name)
	rd.JSON(w, http.StatusCreated, &todo)
}

func (appHandler *AppHandler) removeTodoHandler(w http.ResponseWriter, r *http.Request) {
	sessionId := getSessionID(r)
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	ok := appHandler.db.RemoveTodo(sessionId, id)

	if ok {
		rd.JSON(w, http.StatusOK, Success{true})
	} else {
		rd.JSON(w, http.StatusOK, Success{false})
	}
}

func (appHandler *AppHandler) patchTodoHandler(w http.ResponseWriter, r *http.Request) {
	sessionId := getSessionID(r)
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	revTodo := new(model.Todo)
	err := json.NewDecoder(r.Body).Decode(revTodo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	ok := appHandler.db.PatchTodo(sessionId, id, revTodo.Completed)
	if ok {
		rd.JSON(w, http.StatusOK, Success{true})
	} else {
		rd.JSON(w, http.StatusOK, Success{false})
	}
}

func (appHandler *AppHandler) Close() {
	appHandler.db.Close()
}

func CheckSignin(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// 사용자가 요청한 url이 signin.html일 경우에는 바로 next로 진행한다.
	if strings.Contains(r.URL.Path, "/signin") ||
		strings.Contains(r.URL.Path, "/auth") {
		next(w, r)
		return
	}

	// 사용자가 로그인 되어 있으면 next를 호출하여 다음 체인 핸들러로 진행한다.
	sessionId := getSessionID(r)
	if sessionId != "" {
		next(w, r)
		return
	} else {
		// 만약 사용자가 로그인되어 있지 않다면 Redirect한다.
		http.Redirect(w, r, "/signin.html", http.StatusTemporaryRedirect)
		return
	}
}

func MakeHandler(filepath string) *AppHandler {
	mux := mux.NewRouter()

	// Static전에 로그인 체크를 위하여 Classic을 사용하지 않고 체인순서대로 지정한다.
	// CheckSignin에서 실패할 경우 그 다음 체인인 Newstatic이 동작하지 않음.
	neg := negroni.New(negroni.NewRecovery(), negroni.NewLogger(), negroni.HandlerFunc(CheckSignin), negroni.NewStatic(http.Dir("public")))
	neg.UseHandler(mux)

	app := &AppHandler{
		Handler: neg,
		db:      model.NewDBHandler(filepath),
	}

	mux.HandleFunc("/", app.indexHandler).Methods("GET")
	mux.HandleFunc("/todos", app.getTodoListHandler).Methods("GET")
	mux.HandleFunc("/todos", app.addTodoHandler).Methods("POST")
	mux.HandleFunc("/todos/{id:[0-9]+}", app.removeTodoHandler).Methods("DELETE")
	mux.HandleFunc("/todos/{id:[0-9]+}", app.patchTodoHandler).Methods("PATCH")
	mux.HandleFunc("/auth/google/login", googleLoginHandler)
	mux.HandleFunc("/auth/google/callback", googleAuthCallbackHandler)
	return app
}
