package apiHandler

import (
	"fmt"
	"log"
	"net/http"
	apiService "pet-project/internal/api/service"
	"pet-project/pkg/config"
	"strconv"
	"strings"
)

func EmployeesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		apiService.HandleGetEmployees(w, r)
	case http.MethodPost:
		apiService.HandlePostEmployees(w, r)
	case http.MethodDelete:
		apiService.HandleDeleteEmployees(w, r)
	case http.MethodPut:
		apiService.HandlePutEmployees(w, r)
	}
}
func EmployeeHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	empId := parts[2]
	_, err := strconv.Atoi(empId)
	if err != nil {
		http.Error(w, "Введите корректный ID", http.StatusBadRequest)
		log.Println("Ввели ID не число")
		return
	}
	url := fmt.Sprintf(config.Dbsvc+"/employee/%s", empId)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		http.Error(w, "Ошибка при создании запроса", http.StatusInternalServerError)
		log.Println("Ошибка при создании запроса")
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Ошибка при выполнении запроса", http.StatusBadGateway)
		log.Println("Ошибка при выполнении запроса", err)
		return
	}
	if resp.StatusCode == http.StatusNotFound {
		http.Error(w, "Такого ID нету", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Сотрудник id:%s успешно удален", empId)
}
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	apiService.Login(w, r)
}
