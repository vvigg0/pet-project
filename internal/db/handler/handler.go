package dbHandler

import (
	"database/sql"
	"log"
	"net/http"
	dbService "pet-project/internal/db/service"
	"pet-project/pkg/config"
	"strconv"
	"strings"
)

func EmployeesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		dbService.DbGetEmployees(w, r)
	case http.MethodPost:
		dbService.DbPostEmployees(w, r)
	case http.MethodDelete:
		dbService.DbDeleteEmployees(w, r)
	case http.MethodPut:
		dbService.DbPutEmployees(w, r)
	}
}
func EmployeeHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("postgres", config.PgDsn)
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
