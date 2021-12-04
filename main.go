package main

import (
	"wtfTwitter/database"
	"wtfTwitter/http"
)

type Main struct {
	DB *database.DB
	HTTPServer *http.Server
}

func NewMain() *Main {
	return &Main{
		DB: database.NewDB("host=localhost user=postgres dbname=wtf_twitter port=5432 sslmode=disable"),
		HTTPServer: http.NewServer(),
	}
}

//func Run() {
//
//}

func main() {
	db := database.NewDB("host=localhost user=postgres dbname=wtf_twitter port=5432 sslmode=disable")
	err := database.Open(db)
	if err != nil {
		panic(err)
	}
	server := http.NewServer()
	authService := database.NewAuthService(db)
	server.AuthService = authService
	server.Run(server)
}