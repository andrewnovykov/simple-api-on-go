package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/andrewnovykov/simple-api-on-go/api"
	"github.com/gorilla/mux"
)

type Item struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/items", api.createItem).Methods("POST")
	r.HandleFunc("/items", api.getItems).Methods("GET")
	r.HandleFunc("/items/{id}", api.getItem).Methods("GET")
	r.HandleFunc("/updateitems", api.updateItems).Methods("PUT")
	r.HandleFunc("/register", api.Register).Methods("POST")
	r.HandleFunc("/login", api.Login).Methods("POST")

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	log.Println("Listening on port", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}
