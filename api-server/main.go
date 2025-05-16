package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Employee struct {
	Name       string `json:"name"`
	Secondname string `json:"secondname"`
	Job        string `json:"job"`
	Otdel      int    `json:"otdel"`
}

func addEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Неверный метод", http.StatusMethodNotAllowed)
		return
	}
	var emp Employee
	err := json.NewDecoder(r.Body).Decode(&emp)
	if err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}
	body, _ := json.Marshal(emp)
	resp, err := http.Post("http://dbsvc:8090/addemployee", "application/json", bytes.NewBuffer(body))
	if err != nil || http.StatusCreated != resp.StatusCode {
		http.Error(w, "DB сервер тупанул", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func employeesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Неверный метод", http.StatusMethodNotAllowed)
		return
	}
	query := ""
	if id := r.URL.Query().Get("otdel"); id != "" {
		query = "?otdel=" + id
	}
	resp, err := http.Get("http://dbsvc:8090/employees" + query)
	if err != nil {
		http.Error(w, "Ошибка при выполнении запроса к БД", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	var emps []Employee
	err = json.NewDecoder(resp.Body).Decode(&emps)
	if err != nil {
		http.Error(w, "JSON хуйня", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain;charset=utf-8")
	for i, e := range emps {
		fmt.Fprintf(w, "%d. %s %s %s - отдел %d\n",
			i+1, e.Name, e.Secondname, e.Job, e.Otdel)
	}
}
func main() {
	http.HandleFunc("/employees", employeesHandler)
	http.HandleFunc("/addemployee", addEmployeeHandler)
	log.Println("API сервер запущен на порте 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Не удалось запустить API СЕРВЕР", err)
	}
}
