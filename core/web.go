package core

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

type Todo struct {
	Name      string    `json:"name"`
	Completed bool      `json:"completed"`
	Due       time.Time `json:"due"`
}

type Todos []Todo
func WebUi() {

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", index)
	router.HandleFunc("/todos", TodoIndex)
	router.HandleFunc("/todos/{todoId}", TodoShow)


	s := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	router.PathPrefix("/static").Handler(s)
	fmt.Println("webui监听于8080")

	log.Fatal(http.ListenAndServe(":8080", router))


}

func index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func TodoIndex(w http.ResponseWriter, r *http.Request) {
	todos := Todos{
		Todo{Name: "Write presentation"},
		Todo{Name: "Host meetup"},
	}

	if err := json.NewEncoder(w).Encode(todos); err != nil {
		panic(err)
	}
}

func TodoShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	todoId := vars["todoId"]
	fmt.Fprintf(w, "Todo show: %s\n", todoId)
}