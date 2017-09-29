package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Student struct {
	Id    int
	Name  string
	Range int
}

var db *sql.DB

func errorHandler(w http.ResponseWriter, e error) {
	http.Error(w, e.Error(), http.StatusInternalServerError)
}

func notFoundHandler(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"status": "Not Found"}`))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	rows, err := db.Query("SELECT `id`, `name`, `range` FROM students")
	if err != nil {
		errorHandler(w, err)
	}

	students := make([]Student, 0)

	for rows.Next() {
		s := new(Student)
		err := rows.Scan(&s.Id, &s.Name, &s.Range)
		if err != nil {
			errorHandler(w, err)
		}
		students = append(students, *s)
	}
	json, err := json.Marshal(students)

	if err != nil {
		errorHandler(w, err)
	}

	w.Write(json)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var s Student
	b, _ := ioutil.ReadAll(r.Body)

	json.Unmarshal(b, &s)
	db.Exec("INSERT INTO students(`name`, `range`) VALUES(?, ?)", s.Name, s.Range)
	w.Write([]byte(`{"success": "true"}`))
}

func showHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	s, err := findStudent(params["id"])
	if err != nil {
		notFoundHandler(w)
		return
	}
	json, err := json.Marshal(s)

	if err != nil {
		errorHandler(w, err)
	}
	w.Write(json)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	s, err := findStudent(params["id"])
	if err != nil {
		notFoundHandler(w)
		return
	}

	b, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(b, &s)
	db.Exec("UPDATE students SET `name` = ?, `range` = ? WHERE `id` = ?", s.Name, s.Range, s.Id)
	w.Write([]byte(`{"success": "true"}`))
}

func destroyHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	s, err := findStudent(params["id"])
	if err != nil {
		notFoundHandler(w)
		return
	}
	db.Exec("DELETE FROM students WHERE `id` = ?", s.Id)
	w.Write([]byte(`{"success": "true"}`))
}

func prepareDb() error {
	var err error
	db, _ = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/go_test_db")
	if err = db.Ping(); err != nil {
		return err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS students " +
		"(`id` int(11) NOT NULL AUTO_INCREMENT, `name` VARCHAR(255), " +
		"`range` INT(11), PRIMARY KEY (`id`))")
	if err != nil {
		return err
	}
	return nil
}

func findStudent(id string) (Student, error) {
	var s Student
	err := db.QueryRow("SELECT * FROM students WHERE id = ?", id).Scan(&s.Id, &s.Name, &s.Range)

	if err != nil {
		return Student{}, err
	}
	return s, nil
}

func main() {
	err := prepareDb()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/students", indexHandler).Methods("GET")
	r.HandleFunc("/students", createHandler).Methods("POST")
	r.HandleFunc("/student/{id}", showHandler).Methods("GET")
	r.HandleFunc("/student/{id}", updateHandler).Methods("PATCH")
	r.HandleFunc("/student/{id}", destroyHandler).Methods("DELETE")

	log.Println("Listening on tcp://0.0.0.0:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
