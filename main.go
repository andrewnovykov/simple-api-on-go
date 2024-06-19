package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/andrewnovykov/simple-api-on-go/api"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/items", api.CreateItem).Methods("POST")
	r.HandleFunc("/items", api.GetItems).Methods("GET")
	r.HandleFunc("/items/{id}", api.GetItem).Methods("GET")
	r.HandleFunc("/updateitems", api.UpdateItems).Methods("PUT")
	r.HandleFunc("/register", api.Register).Methods("POST")
	r.HandleFunc("/login", api.Login).Methods("POST")

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	log.Println("Listening on port", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}
