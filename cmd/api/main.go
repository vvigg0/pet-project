package main

import (
	"log"
	"net/http"
	_ "pet-project/docs"
	handler "pet-project/internal/api/handler"
	middleware "pet-project/internal/api/middleware"
	rds "pet-project/internal/api/redis"

	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	rds.Init()
	http.Handle("/employees", middleware.RoleMiddleware(http.HandlerFunc(handler.EmployeesHandler)))
	http.HandleFunc("/login", handler.LoginHandler)
	http.Handle("/swagger/", httpSwagger.WrapHandler)
	log.Println("API сервер запущен на порте 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Не удалось запустить API СЕРВЕР", err)
	}
}
