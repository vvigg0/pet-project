package main

import (
	"log"
	"net/http"
	dbHandler "pet-project/internal/db/handler"
	pg "pet-project/internal/db/postgres"

	_ "github.com/lib/pq"
)

func main() {
	pg.Init()
	http.HandleFunc("/employees", dbHandler.EmployeesHandler)
	log.Println("DB сервер запущен на порте 8090")
	err := http.ListenAndServe(":8090", nil)
	if err != nil {
		log.Fatal("DB сервер не смог запуститься")
	}

}
