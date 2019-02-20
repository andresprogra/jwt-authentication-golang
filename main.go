package main

import (
	"log"
	"net/http"

	"github.com/andresprogra/jwtauth/authentication"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", authentication.Login)
	mux.HandleFunc("/validate", authentication.ValidateToken)

	log.Println("Escuchando en el puerto 8080")

	http.ListenAndServe(":8080", mux)
}
