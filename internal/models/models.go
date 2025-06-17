package models

type AuthResponse struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}
type Employee struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Secondname string `json:"secondname"`
	Job        string `json:"job"`
	Otdel      int    `json:"otdel"`
}
