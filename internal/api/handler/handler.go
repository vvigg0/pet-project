package apiHandler

import (
	"net/http"
	apiService "pet-project/internal/api/service"
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
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	apiService.Login(w, r)
}
