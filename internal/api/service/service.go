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

func isValidWord(s string) error {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return fmt.Errorf("не является валидной строкой")
		}
	}
	return nil
}

func validateQuery(query url.Values) (string, bool) {
	if len(query) == 0 {
		return "", true
	}
	freq := make(map[string]int)
	execStr := []string{}
	fmt.Println(query)
	for k, param := range query {
		if k == "id" {
			for _, id := range param {
				_, err := strconv.Atoi(id)
				if err != nil {
					log.Println("Неверный query параметр ID - ", id)
					continue
				}
				freq[k]++
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
				freq[k]++
				execStr = append(execStr, fmt.Sprintf("name=%s", name))
			}
		} else if k == "secondname" {
			for _, secondname := range param {
				err := isValidWord(secondname)
				if err != nil {
					log.Println("Неверная фамилия - ", secondname)
					continue
				}
				freq[k]++
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
				freq[k]++
				execStr = append(execStr, fmt.Sprintf("job=%s", job))
			}
		} else if k == "otdel" {
			for _, otdel := range param {
				_, err := strconv.Atoi(otdel)
				if err != nil {
					log.Println("Неверный query параметр otdel-", otdel)
					continue
				}
				freq[k]++
				execStr = append(execStr, fmt.Sprintf("otdel=%s", otdel))
			}
		}
	}
	if len(execStr) == 0 {
		return "", false
	}
	validated := "?" + strings.Join(execStr, "&")
	if len(freq) > 2 {
		return validated, false
	}
	for _, v := range freq {
		if v > 1 {
			return validated, false
		}
	}
	return validated, true
}

// HandleGetEmployees godoc
//
//			@Summary	 Вывод сотрудников
//			@Description Выводит сотрудников, отфильтрованных по query string
//			@Description •Если параметров нет - возвращает всех сотрудников.
//			@Description •Можно вводить несколько параметров одного ключа (?id=1&id=2)
//			@Tags		 Employees
//	 		@Produce	 plain
//
//			@Param  Authorization  header string    true   "Authentication header"
//			@Param 	id 			   query  []string  false  "Фильтр: ?id=*" 							collectionFormat(multi)
//			@Param	name		   query  []string  false  "Фильтр: ?name=*"						collectionFormat(multi)
//			@Param	secondname	   query  []string  false  "Фильтр: ?secondname=*"					collectionFormat(multi)
//			@Param	job			   query  []string  false  "Фильтр: ?job=*" (слова разделяются _ )	collectionFormat(multi)
//			@Param	otdel		   query  []string  false  "Фильтр: ?otdel=*"						collectionFormat(multi)
//
//			@Success	200		{string}	string "Список сотрудников построчно"
//			@Success	204		{string}	string "По данному запросу никого нет"
//			@Failure	500 	{string}	string "Ошибка при выполнении запроса к БД  ИЛИ  БД отдало невалидный JSON"
//			@Router		/employees [get]
func HandleGetEmployees(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()
	validQueryStr, shouldCache := validateQuery(urlParams)
	var emps []models.Employee
	start := time.Now()
	if shouldCache {
		val, err := rds.Client.Get(rds.Ctx, "employees"+validQueryStr).Result()
		if err == nil {
			_ = json.Unmarshal([]byte(val), &emps)
			w.Header().Set("Content-Type", "text/plain;charset=utf-8")
			for _, e := range emps {
				fmt.Fprintf(w, "%d. %s %s %s %d\n", e.Id, e.Name, e.Secondname, e.Job, e.Otdel)
			}
			log.Printf("Достали данные из кэша за %v", time.Since(start))
			return
		}
	}
	start = time.Now()
	resp, err := http.Get(config.Dbsvc + "/employees" + validQueryStr)
	if err != nil {
		http.Error(w, "Ошибка при выполнении запроса к БД", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&emps)
	if err != nil {
		http.Error(w, "БД отдало невалидный JSON", http.StatusInternalServerError)
		return
	}
	response, _ := json.Marshal(emps)
	if shouldCache && len(emps) != 0 {
		rds.Client.Set(rds.Ctx, "employees"+validQueryStr, response, time.Minute)
		log.Println("Занесли в кэш запрос: " + "employees" + validQueryStr)
	}
	if len(emps) > 0 {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain;charset=utf-8")
		for _, e := range emps {
			fmt.Fprintf(w, "%d. %s %s %s %d\n", e.Id, e.Name, e.Secondname, e.Job, e.Otdel)
		}
	} else {
		w.WriteHeader(http.StatusNoContent)
		w.Header().Set("Content-Type", "text/plain;charset=utf-8")
		fmt.Fprintf(w, "По данному запросу никого нет")
	}
	log.Printf("Достали данные из БД за %v", time.Since(start))
}

// HandlePostEmployees godoc
//
//	@Summary	 Добавление сотрудников
//	@Description Добавляет сотрудников в БД(все поля JSON должны быть заполенны)
//	@Tags		 Employees
//	@Accept		 json
//	@Produce	 plain
//
//	@Param  	 Authorization  header 	string 				true 	"Authentication header"
//	@Param		 employees	    body	[]models.Employee	true	"Массив сотрудников"
//
//	@Success	 201	{string}	string	"Успешно добавлены сотрудники: ... ..."
//	@Failure	 400	{string}	string	"Неверный json"
//	@Failure 	 500	{string}	string	"Ошибка при выполнении запроса к БД"
//	@Router		 /employees [post]
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
		http.Error(w, "Ошибка при выполнении запроса к БД", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	var names string
	for _, emp := range emps {
		names = names + " " + emp.Name
	}
	fmt.Fprintf(w, "Успешно добавлены сотрудники: %s", names)
	clearCache(emps)
}
func findKeys(keys []string) []string {
	result := []string{}
	var cursor uint64
	for i := 0; i < len(keys); i++ {
		keyValue := strings.Split(keys[i], "=")
		pattern := fmt.Sprintf("*%s=%v*", keyValue[0], keyValue[1])
		for {
			keys, newCursor, err := rds.Client.Scan(rds.Ctx, cursor, pattern, 100).Result()
			if err != nil {
				log.Println("Ошибка сканирования Redis: ", err)
			}
			result = append(result, keys...)
			cursor = newCursor
			if cursor == 0 {
				break
			}
		}
	}
	return result
}

// HandleDeleteEmployees godoc
//
//		@Summary	 Удаление сотрудников
//		@Description Удаляет сотрудников, отфильтрованных по query string
//		@Description •Если параметров нет - удаляет **всех** сотрудников (опасная операция!)
//		@Description •Можно вводить несколько параметров одного ключа (?id=1&id=2)
//		@Tags		 Employees
//	 	@Produce	 plain
//
//		@Param  Authorization  header string    true   "Authentication header"
//		@Param 	id 			   query  []string  false  "Фильтр: ?id=*" 							collectionFormat(multi)
//		@Param	name		   query  []string  false  "Фильтр: ?name=*"						collectionFormat(multi)
//		@Param	secondname	   query  []string  false  "Фильтр: ?secondname=*"					collectionFormat(multi)
//		@Param	job			   query  []string  false  "Фильтр: ?job=*" (слова разделяются _ )	collectionFormat(multi)
//		@Param	otdel		   query  []string  false  "Фильтр: ?otdel=*"						collectionFormat(multi)
//
//		@Success 201 {string} string "Сотрудники удалены"
//		@Failure 500 {string} string "Ошибка при создании запроса"
//		@Failure 502 {string} string "Ошибка при выполнении запроса"
//		@Router /employees [delete]
func HandleDeleteEmployees(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()
	validQueryStr, shouldCache := validateQuery(urlParams)
	var emps []models.Employee
	if shouldCache {
		var err error
		emps, err = fetchForInvalidation(validQueryStr)
		if err != nil {
			log.Println("Не удалось получить сотрудников для удаления кэша")
		}
	}
	req, err := http.NewRequest("DELETE", config.Dbsvc+"/employees"+validQueryStr, nil)
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
	if shouldCache {
		clearCache(emps)
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Сотрудники удалены")
}

// HandlePutEmployees godoc
//
//		@Summary	 Изменение сотрудников
//		@Description Изменяет сотрудников, отфильтрованных по query string,меняя их данные на данные из JSON
//		@Description •Если параметров нет - изменяет **всех** сотрудников (опасная операция!)
//		@Description •Можно вводить несколько параметров одного ключа (?id=1&id=2)
//		@Tags		 Employees
//		@Accept		 json
//	 	@Produce	 plain
//
//		@Param  Authorization  header string    		true   	"Authentication header"
//		@Param 	id 			   query  []string  		false  	"Фильтр: ?id=*" 							collectionFormat(multi)
//		@Param	name		   query  []string  		false  	"Фильтр: ?name=*"						collectionFormat(multi)
//		@Param	secondname	   query  []string  		false  	"Фильтр: ?secondname=*"					collectionFormat(multi)
//		@Param	job			   query  []string  		false  	"Фильтр: ?job=*" (слова разделяются _ )	collectionFormat(multi)
//		@Param	otdel		   query  []string  		false  	"Фильтр: ?otdel=*"						collectionFormat(multi)
//		@Param  data		   body	  models.Employee	true	"Данные на которые надо поменять"
//
//		@Success  201 {string}  string  "Изменение выполнено"
//		@Failure  400 {string}	string  "Неверный JSON"
//		@Failure 500 {string} string "Ошибка при создании запроса"
//		@Failure 502 {string} string "Ошибка при выполнении запроса"
//		@Router /employees [put]
func HandlePutEmployees(w http.ResponseWriter, r *http.Request) {
	var emp models.Employee
	err := json.NewDecoder(r.Body).Decode(&emp)
	if err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}
	body, _ := json.Marshal(emp)
	urlParams := r.URL.Query()
	validQueryStr, shouldCache := validateQuery(urlParams)
	var emps []models.Employee
	if shouldCache {
		var err error
		emps, err = fetchForInvalidation(validQueryStr)
		if err != nil {
			log.Println("Не удалось получить сотрудников для удаления кэша")
		}
	}
	req, err := http.NewRequest("PUT", config.Dbsvc+"/employees"+validQueryStr, bytes.NewBuffer(body))
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
	clearCache(emps)
	defer resp.Body.Close()
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, "Изменение выполнено")
}
func fetchForInvalidation(queryStr string) ([]models.Employee, error) {
	resp, err := http.Get(config.Dbsvc + "/employees" + queryStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var emps []models.Employee
	if err := json.NewDecoder(resp.Body).Decode(&emps); err != nil {
		return nil, err
	}
	return emps, nil
}
func clearCache(emps []models.Employee) {
	keys := make(map[string]int)
	var uniqKeys []string
	for _, emp := range emps {
		keys[fmt.Sprintf("id=%v", emp.Id)] = 0
		keys[fmt.Sprintf("name=%v", emp.Name)] = 0
		keys[fmt.Sprintf("secondname=%v", emp.Secondname)] = 0
		keys[fmt.Sprintf("job=%v", emp.Job)] = 0
		keys[fmt.Sprintf("otdel=%v", emp.Otdel)] = 0
	}
	_, _ = rds.Client.Del(rds.Ctx, "employees").Result()
	for k := range keys {
		uniqKeys = append(uniqKeys, k)
	}
	confKeys := findKeys(uniqKeys)
	for _, val := range confKeys {
		_, err := rds.Client.Del(rds.Ctx, val).Result()
		if err != nil {
			continue
		}
	}
	log.Println("Записи из кэша удалены успешно")
}

// Login godoc
//
// @Summary		Вход для получения JWT токена
// @Description Дает токен+роль,роль админа выдается при вводе данных админа из .env файла
// @Tags 		Authorization
// @Accept		json
// @Produce		json
//
// @Param credentials body object{username=string,password=string} true "Данные для входа"
// @Success 201 {object} models.AuthResponse "JWT+роль"
// @Failure 400 {string} string "Неверные данные"
// @Failure 500 {string} string "Ошибка генерации токена"
// @Router	/login [post]
func Login(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Неверные данные", http.StatusBadRequest)
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
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.AuthResponse{Token: signedToken, Role: role})
}
