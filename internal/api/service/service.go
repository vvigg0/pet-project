package apiService

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	rds "pet-project/internal/api/redis"
	"pet-project/internal/models"
	"pet-project/pkg/config"
	"pet-project/pkg/myJwt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func IsValidWord(s string) error {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return fmt.Errorf("не является валидной строкой")
		}
	}
	return nil
}

func ValidateQuery(query url.Values) string {
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

func HandleGetEmployees(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()
	legitQueryStr := ValidateQuery(urlParams)
	cacheKey := "employees" + legitQueryStr
	var emps []models.Employee
	start := time.Now()
	val, err := rds.Client.Get(rds.Ctx, cacheKey).Result()
	if err == nil {
		_ = json.Unmarshal([]byte(val), &emps)
		w.Header().Set("Content-Type", "text/plain;charset=utf-8")
		for _, e := range emps {
			fmt.Fprintf(w, "%d. %s %s %s %d\n", e.Id, e.Name, e.Secondname, e.Job, e.Otdel)
		}
		log.Printf("Достали данные из кэша за %v", time.Since(start))
		return
	}
	start = time.Now()
	resp, err := http.Get(config.Dbsvc + "/employees" + legitQueryStr)
	if err != nil {
		http.Error(w, "Ошибка при выполнении запроса к БД", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	log.Println(resp.Body)
	err = json.NewDecoder(resp.Body).Decode(&emps)
	if err != nil {
		http.Error(w, "JSON хуйня", http.StatusInternalServerError)
		return
	}
	response, _ := json.Marshal(emps)
	rds.Client.Set(rds.Ctx, cacheKey, response, time.Minute*10)
	w.Header().Set("Content-Type", "text/plain;charset=utf-8")
	for _, e := range emps {
		fmt.Fprintf(w, "%d. %s %s %s %d\n", e.Id, e.Name, e.Secondname, e.Job, e.Otdel)
	}
	log.Printf("Достали данные из БД за %v", time.Since(start))
}
func HandlePostEmployees(w http.ResponseWriter, r *http.Request) {
	var emps []models.Employee
	body, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(body, &emps)
	if err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}
	body, _ = json.Marshal(emps)
	resp, err := http.Post(config.Dbsvc+"/employees", "application/json", bytes.NewBuffer(body))
	if err != nil || http.StatusCreated != resp.StatusCode {
		http.Error(w, "DB сервер тупанул", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	var names string
	for _, emp := range emps {
		names = names + " " + emp.Name
	}
	fmt.Fprintf(w, "Успешно добавлены сотрудники: %s", names)
}
func HandleDeleteEmployees(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()
	queryStr := ValidateQuery(urlParams)
	req, err := http.NewRequest("DELETE", config.Dbsvc+"/employees"+queryStr, nil)
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
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Сотрудники удалены")
}
func HandlePutEmployees(w http.ResponseWriter, r *http.Request) {
	var emp models.Employee
	err := json.NewDecoder(r.Body).Decode(&emp)
	if err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}
	body, _ := json.Marshal(emp)
	urlParams := r.URL.Query()
	legitQueryStr := ValidateQuery(urlParams)
	req, err := http.NewRequest("PUT", config.Dbsvc+"/employees"+legitQueryStr, bytes.NewBuffer(body))
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
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, "Изменение выполнено")
}
func Login(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}
	var role string
	switch {
	case creds.Username == config.AdminName && creds.Password == config.AdminPassword:
		role = "admin"
	default:
		role = "guest"
	}
	signedToken, err := myJwt.SignToken(creds.Username, role)
	if err != nil {
		http.Error(w, "Ошибка генерации токена", http.StatusInternalServerError)
		log.Printf("Ошибка генерации токена: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": signedToken})
	fmt.Fprintf(w, "Ваша роль: %s", role)
}
