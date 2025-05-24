package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
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
	result, err := db.Exec("Insert into sotrudniki(id,name,secondname,job,otdel) VALUES($1,$2,$3,$4,$5)",
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
	givenQuery := r.URL.Query()
	query := []string{}
	execQuery := ""
	count := 0
	if len(givenQuery) > 0 {
		for k := range givenQuery {
			log.Println(k)
			if len(givenQuery[k]) == 1 {
				for _, val := range givenQuery[k] {
					count++
					s := []byte(val)
					for i, l := range s {
						if l == 95 {
							s[i] = 32
							val = string(s)
						}
					}
					query = append(query, fmt.Sprintf("%s=$%d", k, count))
					args = append(args, val)
				}
			} else {
				subQuery := fmt.Sprintf(" %s IN ", k)
				placeholders := []string{}
				for _, vals := range givenQuery[k] {
					vals := strings.Split(vals, " ")
					for _, val := range vals {
						count++
						s := []byte(val)
						for i, l := range s {
							if l == 95 {
								s[i] = 32
								val = string(s)
							}
						}
						placeholders = append(placeholders, fmt.Sprintf("$%d", count))
						if k == "id" || k == "otdel" {
							intVal, _ := strconv.Atoi(val)
							args = append(args, intVal)
						} else {
							args = append(args, val)
						}
					}
				}
				query = append(query, subQuery+"("+strings.Join(placeholders, ",")+")")
			}
		}
		execQuery = " WHERE " + strings.Join(query, " AND ")
	} else {
		execQuery = ""
	}
	log.Println(args)
	rows, err := db.Query("SELECT id,name,secondname,job,otdel from sotrudniki"+execQuery, args...)
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
