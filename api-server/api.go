package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	rds "servers/pkg/cache"
	"strconv"
	"strings"
	"time"
	"unicode"

	jwt "github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte(os.Getenv("SECRET_KEY"))
var dbsvc string = os.Getenv("DB_SVC_URL")

type Employee struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Secondname string `json:"secondname"`
	Job        string `json:"job"`
	Otdel      int    `json:"otdel"`
}
type CustomClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func isValidWord(s string) error {
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
				err := isValidWord(name)
				if err != nil {
					log.Println("Неверное имя - ", name)
					continue
				}
				execStr = append(execStr, fmt.Sprintf("name=%s", name))
			}
		} else if k == "secondname" {
			for _, secondname := range param {
				err := isValidWord(secondname)
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
					err := isValidWord(part)
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
func handleGetEmployees(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()
	legitQueryStr := validateQuery(urlParams)
	cacheKey := "employees" + legitQueryStr
	var emps []Employee
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
	resp, err := http.Get(dbsvc + "/employees" + legitQueryStr)
	if err != nil {
		http.Error(w, "Ошибка при выполнении запроса к БД", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
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
func handlePostEmployees(w http.ResponseWriter, r *http.Request) {
	var emps []Employee
	body, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(body, &emps)
	if err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}
	body, _ = json.Marshal(emps)
	resp, err := http.Post(dbsvc+"/employees", "application/json", bytes.NewBuffer(body))
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
func handleDeleteEmployees(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()
	queryStr := validateQuery(urlParams)
	req, err := http.NewRequest("DELETE", dbsvc+"/employees"+queryStr, nil)
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
func handlePutEmployees(w http.ResponseWriter, r *http.Request) {
	var emp Employee
	err := json.NewDecoder(r.Body).Decode(&emp)
	if err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}
	body, _ := json.Marshal(emp)
	urlParams := r.URL.Query()
	legitQueryStr := validateQuery(urlParams)
	req, err := http.NewRequest("PUT", dbsvc+"/employees"+legitQueryStr, bytes.NewBuffer(body))
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
func employeesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetEmployees(w, r)
	case http.MethodPost:
		handlePostEmployees(w, r)
	case http.MethodDelete:
		handleDeleteEmployees(w, r)
	case http.MethodPut:
		handlePutEmployees(w, r)
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
	url := fmt.Sprintf(dbsvc+"/employee/%s", empId)
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
func loginHandler(w http.ResponseWriter, r *http.Request) {
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
	case creds.Username == os.Getenv("ADMIN_NAME") && creds.Password == os.Getenv("ADMIN_PASSWORD"):
		role = "admin"
	default:
		role = "guest"
	}
	claims := CustomClaims{
		Username: creds.Username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			Issuer:    "jwt-server",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secretKey)
	if err != nil {
		http.Error(w, "Ошибка генерации токена", http.StatusInternalServerError)
		log.Printf("Ошибка генерации токена: %v,%v", token, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": signedToken})
	fmt.Fprintf(w, "Ваша роль: %s", claims.Role)
}
func RoleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "Нет токена", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Неверный токен", http.StatusUnauthorized)
			log.Printf("Неверный токен: %v,%v", token, err)
			return
		}
		role := token.Claims.(*CustomClaims).Role
		switch r.Method {
		case "POST", "PUT", "DELETE":
			if role != "admin" {
				http.Error(w, "Недостаточно прав", http.StatusForbidden)
				return
			}
		default:
		}
		next.ServeHTTP(w, r)
	})
}
func main() {
	rds.Init()
	http.Handle("/employees", RoleMiddleware(http.HandlerFunc(employeesHandler)))
	http.Handle("/employee/", RoleMiddleware(http.HandlerFunc(employeeHandler)))
	http.HandleFunc("/login", loginHandler)
	log.Println("API сервер запущен на порте 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Не удалось запустить API СЕРВЕР", err)
	}
}
