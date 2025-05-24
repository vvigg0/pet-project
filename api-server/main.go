package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode"
)

type Employee struct {
	Id         int    `json:"id"`
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
func IsValidWord(s string) error {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return fmt.Errorf("не является валидной строкой")
		}
	}
	return nil
}
func validateQuery(query url.Values) string {
	execStr := []string{}
	for k, param := range query {
		if k == "id" {
			for _, id := range param {
				_, err := strconv.Atoi(id)
				if err != nil {
					log.Println("Неверный query параметр ID - ", id)
					continue
				}
				execStr = append(execStr, fmt.Sprintf("id=%s", id))
			}
			break
		} else if k == "name" {
			for _, name := range param {
				err := IsValidWord(name)
				if err != nil {
					log.Println("Неверное имя - ", name)
					continue
				}
				execStr = append(execStr, fmt.Sprintf("name=%s", name))
			}
		} else if k == "secondname" {
			for _, secondname := range param {
				err := IsValidWord(secondname)
				if err != nil {
					log.Println("Неверная фамилия - ", secondname)
					continue
				}
				execStr = append(execStr, fmt.Sprintf("secondname=%s", secondname))
			}
		} else if k == "job" {
			for _, job := range param {
				partsJob := strings.Split(job, "_")
				for _, part := range partsJob {
					err := IsValidWord(part)
					if err != nil {
						log.Println("Неверная должность - ", job)
						continue
					}
				}
				execStr = append(execStr, fmt.Sprintf("job=%s", job))
			}
		} else if k == "otdel" {
			for _, otdel := range param {
				_, err := strconv.Atoi(otdel)
				if err != nil {
					log.Println("Неверный query параметр otdel-", otdel)
					continue
				}
				execStr = append(execStr, fmt.Sprintf("otdel=%s", otdel))
			}
		}
	}
	if len(execStr) == 0 {
		return ""
	}
	return "?" + strings.Join(execStr, "&")
}
func employeesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Неверный метод", http.StatusMethodNotAllowed)
		return
	}
	urlParams := r.URL.Query()
	legitQueryStr := validateQuery(urlParams)
	log.Println(legitQueryStr)
	resp, err := http.Get("http://dbsvc:8090/employees" + legitQueryStr)
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
	for _, e := range emps {
		fmt.Fprintf(w, "%d. %s %s %s - отдел %d\n",
			e.Id, e.Name, e.Secondname, e.Job, e.Otdel)
	}
}
func employeeHandler(w http.ResponseWriter, r *http.Request) {
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
	url := fmt.Sprintf("http://dbsvc:8090/employee/%s", empId)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		http.Error(w, "Ошибка при создании запроса", http.StatusInternalServerError)
		log.Println("Ошибка при создании запроса")
		return
	}
	resp, err := http.DefaultClient.Do(req)
	log.Println(err)
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
func main() {
	http.HandleFunc("/employees", employeesHandler)
	http.HandleFunc("/addemployee", addEmployeeHandler)
	http.HandleFunc("/employee/", employeeHandler)
	log.Println("API сервер запущен на порте 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Не удалось запустить API СЕРВЕР", err)
	}
}
