package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "12345"
	dbname   = "dbgolang"
)

var psqlInfo string = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
	host, port, user, password, dbname)
var dsn string = os.Getenv("PG_DSN")

type Employee struct {
	Name       string `json:"name"`
	Secondname string `json:"secondname"`
	Job        string `json:"job"`
	Otdel      int    `json:"otdel"`
}

func addEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		http.Error(w, "Возникли проблемы на стороне БД", 500)
		return
	}
	defer db.Close()
	var emp Employee
	err = json.NewDecoder(r.Body).Decode(&emp)
	if err != nil {
		http.Error(w, "Хуйня JSON", http.StatusBadRequest)
		return
	}
	result, err := db.Exec("Insert into sotrudniki(имя,фамилия,должность,отдел_id) VALUES($1,$2,$3,$4)",
		emp.Name, emp.Secondname, emp.Job, emp.Otdel)
	if err != nil {
		http.Error(w, "Неизвестная ошибка при работе с БД", 520)
		return
	}
	rowsAffected, _ := result.RowsAffected()
	w.WriteHeader(http.StatusCreated)
	log.Printf("Успешно вставлен %d сотрудник", rowsAffected)
}
func employeesHandler(w http.ResponseWriter, r *http.Request) {
	employees := []Employee{}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		http.Error(w, "Возникла проблема при подключении к БД", http.StatusInternalServerError)
		log.Println("Возникла проблема при подключении к БД")
		return
	}
	defer db.Close()
	var args []interface{}
	query := ""
	if otdel := r.URL.Query().Get("otdel"); otdel != "" {
		query = " WHERE отдел_id=$1"
		args = append(args, otdel)
	}
	rows, err := db.Query("SELECT имя,фамилия,должность,отдел_id from sotrudniki"+query, args...)
	if err != nil {
		http.Error(w, "Возникла ошибка при выполнении очевидного запроса", http.StatusInternalServerError)
		log.Println("Ошибка при запросе сотрудников", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		emp := Employee{}
		err := rows.Scan(&emp.Name, &emp.Secondname, &emp.Job, &emp.Otdel)
		if err != nil {
			log.Println("Возникла ошибка при обработке одного из сотрудников")
		}
		employees = append(employees, emp)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(employees)
}
func main() {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Не удалось подключиться к БД")
	}
	db.Close()
	http.HandleFunc("/addemployee", addEmployeeHandler)
	http.HandleFunc("/employees", employeesHandler)
	log.Println("DB сервер запущен на порте 8090")
	err = http.ListenAndServe(":8090", nil)
	if err != nil {
		log.Fatal("DB сервер не смог запуститься")
	}

}
