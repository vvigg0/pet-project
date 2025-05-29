package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
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

func buildGetExecQuery(givenQuery url.Values, count int) (string, []interface{}) {
	var args []interface{}
	query := []string{}
	cnt := count
	execQuery := ""
	if len(givenQuery) > 0 {
		for k := range givenQuery {
			log.Println(k)
			if len(givenQuery[k]) == 1 {
				for _, val := range givenQuery[k] {
					cnt++
					s := []byte(val)
					for i, l := range s {
						if l == 95 {
							s[i] = 32
							val = string(s)
						}
					}
					query = append(query, fmt.Sprintf("%s=$%d", k, cnt))
					args = append(args, val)
				}
			} else {
				subQuery := fmt.Sprintf(" %s IN ", k)
				placeholders := []string{}
				for _, vals := range givenQuery[k] {
					vals := strings.Split(vals, " ")
					for _, val := range vals {
						cnt++
						s := []byte(val)
						for i, l := range s {
							if l == 95 {
								s[i] = 32
								val = string(s)
							}
						}
						placeholders = append(placeholders, fmt.Sprintf("$%d", cnt))
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
	return execQuery, args
}
func dbGetEmployees(w http.ResponseWriter, r *http.Request) {
	employees := []Employee{}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		http.Error(w, "Возникла проблема при подключении к БД", http.StatusInternalServerError)
		log.Println("Возникла проблема при подключении к БД")
		return
	}
	defer db.Close()
	givenQuery := r.URL.Query()
	execQuery, args := buildGetExecQuery(givenQuery, 0)
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
func buildInsertQuery(emps []Employee) (string, []interface{}) {
	count := 0
	i := 0
	var args []interface{}
	resultPlHolds := []string{}
	for _, emp := range emps {
		i++
		placeholders := []string{}
		v := reflect.ValueOf(emp)
		size := v.NumField()
		for count < size*i {
			count++
			placeholders = append(placeholders, fmt.Sprintf("$%d", count))
		}
		for i := 0; i < size; i++ {
			args = append(args, v.Field(i).Interface())
		}
		resultPlHolds = append(resultPlHolds, "("+strings.Join(placeholders, ",")+")")
	}
	placeholdStr := strings.Join(resultPlHolds, ",")
	return placeholdStr, args
}
func dbPostEmployees(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		http.Error(w, "Возникли проблемы на стороне БД", 500)
		return
	}
	defer db.Close()
	var emps []Employee
	_ = json.NewDecoder(r.Body).Decode(&emps)
	placeholders, args := buildInsertQuery(emps)
	result, err := db.Exec("Insert into sotrudniki(id,name,secondname,job,otdel) VALUES"+placeholders, args...)
	if err != nil {
		http.Error(w, "Неизвестная ошибка при работе с БД", 520)
		log.Printf(placeholders, args...)
		log.Println("Ошибка в билде запроса")
		return
	}
	rowsAffected, _ := result.RowsAffected()
	w.WriteHeader(http.StatusCreated)
	log.Printf("Успешно вставлен %d сотрудник", rowsAffected)
}
func dbDeleteEmployees(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		http.Error(w, "Возникли проблемы при открытии базы данных", 500)
		return
	}
	defer db.Close()
	urlQuery := r.URL.Query()
	execQuery, args := buildGetExecQuery(urlQuery, 0)
	result, err := db.Exec("DELETE FROM sotrudniki"+execQuery, args...)
	if err != nil {
		http.Error(w, "Не удалось выполнить запрос", http.StatusBadGateway)
		log.Println("Возникла проблема при выполнении запроса", execQuery, args)
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Не нашлось ни одного сотрудника,который соответствует такому условию", http.StatusNotFound)
		log.Println("Никого нету по такому условию удаления")
		return
	}
	log.Println("Удаление прошло успешно")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Удаление прошло успешно")
}
func buildPutQuery(urlquery url.Values, emp Employee) (string, []interface{}) {
	count := 0
	SetQuery := []string{}
	var args []interface{}
	if emp.Id != 0 {
		count++
		SetQuery = append(SetQuery, fmt.Sprintf("id=$%d", count))
		args = append(args, emp.Id)
	}
	if emp.Name != "" {
		count++
		SetQuery = append(SetQuery, fmt.Sprintf("name=$%d", count))
		args = append(args, emp.Name)
	}
	if emp.Secondname != "" {
		count++
		SetQuery = append(SetQuery, fmt.Sprintf("secondname=$%d", count))
		args = append(args, emp.Secondname)
	}
	if emp.Job != "" {
		count++
		SetQuery = append(SetQuery, fmt.Sprintf("job=$%d", count))
		args = append(args, emp.Job)
	}
	if emp.Otdel != 0 {
		count++
		SetQuery = append(SetQuery, fmt.Sprintf("otdel=$%d", count))
		args = append(args, emp.Otdel)
	}
	if len(SetQuery) == 0 {
		log.Println("Неверный PUT запрос")
	}
	resultQuery := strings.Join(SetQuery, ",")
	WhereQuery, args2 := buildGetExecQuery(urlquery, count)
	args = append(args, args2...)
	resultQuery = resultQuery + WhereQuery
	return resultQuery, args
}
func dbPutEmployees(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		http.Error(w, "Не удалось подключиться к базе данных", http.StatusInternalServerError)
		log.Println("Не удалось подключиться к БД")
		return
	}
	var emp Employee
	_ = json.NewDecoder(r.Body).Decode(&emp)
	urlQuery := r.URL.Query()
	execQuery, args := buildPutQuery(urlQuery, emp)
	result, err := db.Exec("UPDATE sotrudniki SET "+execQuery, args...)
	log.Printf(execQuery, args...)
	if err != nil {
		http.Error(w, "Не удалось выполнить запрос", http.StatusBadGateway)
		log.Println("Возникла проблема при выполнении запроса"+execQuery, args)
		log.Println(err)
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Не нашлось ни одного сотрудника,который соответствует такому условию", http.StatusNotFound)
		log.Println("Никого нету по такому условию изменения")
		return
	}
	log.Println("Изменение прошло успешно")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Изменение прошло успешно")
}
func employeesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		dbGetEmployees(w, r)
	case http.MethodPost:
		dbPostEmployees(w, r)
	case http.MethodDelete:
		dbDeleteEmployees(w, r)
	case http.MethodPut:
		dbPutEmployees(w, r)
	}
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
	http.HandleFunc("/employees", employeesHandler)
	http.HandleFunc("/employee/", employeeHandler)
	log.Println("DB сервер запущен на порте 8090")
	err = http.ListenAndServe(":8090", nil)
	if err != nil {
		log.Fatal("DB сервер не смог запуститься")
	}

}
