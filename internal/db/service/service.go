package dbService

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"pet-project/internal/models"
	"pet-project/pkg/config"
	"strconv"
	"strings"
)

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

func DbGetEmployees(w http.ResponseWriter, r *http.Request) {
	employees := []models.Employee{}
	log.Println(config.PgDsn)
	db, err := sql.Open("postgres", config.PgDsn)
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
		emp := models.Employee{}
		err := rows.Scan(&emp.Id, &emp.Name, &emp.Secondname, &emp.Job, &emp.Otdel)
		if err != nil {
			log.Println("Возникла ошибка при обработке одного из сотрудников")
		}
		employees = append(employees, emp)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(employees)
}
func buildInsertQuery(emps []models.Employee) (string, []interface{}) {
	count := 0
	var args []interface{}
	placeholders := []string{}
	resultPlHolds := []string{}
	for _, emp := range emps {
		placeholders = append(placeholders, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d)", count+1, count+2, count+3, count+4, count+5))
		args = append(args, emp.Id, emp.Name, emp.Secondname, emp.Job, emp.Otdel)
		count += 5
	}
	resultPlHolds = append(resultPlHolds, placeholders...)
	placeholdStr := strings.Join(resultPlHolds, ",")
	return placeholdStr, args
}
func DbPostEmployees(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("postgres", config.PgDsn)
	if err != nil {
		http.Error(w, "Возникли проблемы на стороне БД", 500)
		return
	}
	defer db.Close()
	var emps []models.Employee
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
func DbDeleteEmployees(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("postgres", config.PgDsn)
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
func buildPutQuery(urlquery url.Values, emp models.Employee) (string, []interface{}) {
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
func DbPutEmployees(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("postgres", config.PgDsn)
	if err != nil {
		http.Error(w, "Не удалось подключиться к базе данных", http.StatusInternalServerError)
		log.Println("Не удалось подключиться к БД")
		return
	}
	var emp models.Employee
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
