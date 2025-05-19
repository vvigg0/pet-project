package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

var dsn string = os.Getenv("PG_DSN")

type Employee struct {
	Id         int    `json:"id"`
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
	result, err := db.Exec("Insert into sotrudniki(id,имя,фамилия,должность,отдел_id) VALUES($1,$2,$3,$4,$5)",
		emp.Id, emp.Name, emp.Secondname, emp.Job, emp.Otdel)
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
	rows, err := db.Query("SELECT id,имя,фамилия,должность,отдел_id from sotrudniki"+query, args...)
	if err != nil {
		http.Error(w, "Возникла ошибка при выполнении очевидного запроса", http.StatusInternalServerError)
		log.Println("Ошибка при запросе сотрудников", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		emp := Employee{}
		err := rows.Scan(&emp.Id, &emp.Name, &emp.Secondname, &emp.Job, &emp.Otdel)
		if err != nil {
			log.Println("Возникла ошибка при обработке одного из сотрудников")
		}
		employees = append(employees, emp)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(employees)
}
func employeeHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		http.Error(w, "Ошибка при открытии БД", http.StatusInternalServerError)
		log.Println("Ошибка при открытии базы данных")
		return
	}
	defer db.Close()
	parts := strings.Split(r.URL.Path, "/")
	empId := parts[2]
	intEmpId, _ := strconv.Atoi(empId)
	if intEmpId < 1 {
		http.Error(w, "Неверный ID", http.StatusNotFound)
		log.Println("Ввели неверный ID")
		return
	}
	result, err := db.Exec("DELETE FROM sotrudniki WHERE id=$1", intEmpId)
	if err != nil {
		http.Error(w, "Ошибка при удалении сотрудника", http.StatusInternalServerError)
		log.Println("Ошибка при удалении сотрудника")
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Такого ID нету,ничего не удалилось", http.StatusNotFound)
		log.Println("Несуществующий ID,ничего не удалилось")
		return
	}
	log.Println("Сотрудник успешно удален")
	w.WriteHeader(http.StatusCreated)
}
func main() {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Не удалось подключиться к БД")
	}
	db.Close()
	http.HandleFunc("/addemployee", addEmployeeHandler)
	http.HandleFunc("/employees", employeesHandler)
	http.HandleFunc("/employee/", employeeHandler)
	log.Println("DB сервер запущен на порте 8090")
	err = http.ListenAndServe(":8090", nil)
	if err != nil {
		log.Fatal("DB сервер не смог запуститься")
	}

}
