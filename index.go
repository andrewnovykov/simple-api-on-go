package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

type Item struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

var (
	store  = make(map[int]Item)
	mu     sync.RWMutex
	nextID = 1
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/items", createItem).Methods("POST")
	r.HandleFunc("/items", getItems).Methods("GET")
	r.HandleFunc("/items/{id}", getItem).Methods("GET")

	port := ":8080"
	fmt.Printf("Server running on port%s\n", port)
	http.ListenAndServe(port, r)
}

func createItem(w http.ResponseWriter, r *http.Request) {
	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	item.ID = nextID
	store[item.ID] = item
	nextID++
	mu.Unlock()

	// Read the existing items from the file
	var items []Item
	file, err := os.Open("items.json")
	if err != nil {
		if !os.IsNotExist(err) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err := json.NewDecoder(file).Decode(&items); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		file.Close()
	}

	// Append the new item to the slice
	items = append(items, item)

	// Write the slice back to the file
	file, err = os.Create("items.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(items); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

func getItems(w http.ResponseWriter, r *http.Request) {
	// Open the file
	file, err := os.Open("items.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Decode the items from the file
	var items []Item
	if err := json.NewDecoder(file).Decode(&items); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Encode the items to the response
	json.NewEncoder(w).Encode(items)
}

func getItem(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid id: %s", idStr), http.StatusBadRequest)
		return
	}

	// Open the file
	file, err := os.Open("items.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Decode the items from the file
	var items []Item
	if err := json.NewDecoder(file).Decode(&items); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Find the item with the given ID
	var item *Item
	for _, i := range items {
		if i.ID == id {
			item = &i
			break
		}
	}

	if item == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("no item with id: %d", id)})
		return
	}

	json.NewEncoder(w).Encode(item)
}
