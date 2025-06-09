package config

import "os"

var Dbsvc string = os.Getenv("DB_SVC_URL")
var AdminName string = os.Getenv("ADMIN_NAME")
var AdminPassword string = os.Getenv("ADMIN_PASSWORD")
var SecretKey = []byte(os.Getenv("SECRET_KEY"))
var PgDsn string = os.Getenv("PG_DSN")
