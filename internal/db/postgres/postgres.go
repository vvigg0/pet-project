package pg

import (
	"database/sql"
	"log"
	"pet-project/pkg/config"
)

func Init() {
	db, err := sql.Open("postgres", config.PgDsn)
	if err != nil {
		log.Fatal("Не удалось подключиться к БД")
	}
	db.Close()
}
