package dbHandler

import (
	"net/http"
	dbService "pet-project/internal/db/service"
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
