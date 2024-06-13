package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type Item struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Token    string `json:"token,omitempty"`
}

type Token struct {
	Token string `json:"token"`
}

var (
	store = make(map[int]Item)
	users = make(map[int]User)
	mu    sync.RWMutex
)

const (
	errBadRequest     = "bad request"
	errUnauthorized   = "unauthorized"
	errNotFound       = "not found"
	errInternalServer = "internal server error"
	errMissingHeader  = "missing authorization header"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/items", createItem).Methods("POST")
	r.HandleFunc("/items", getItems).Methods("GET")
	r.HandleFunc("/items/{id}", getItem).Methods("GET")
	r.HandleFunc("/updateitems", updateItems).Methods("PUT")
	r.HandleFunc("/register", register).Methods("POST")
	r.HandleFunc("/login", login).Methods("POST")

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	log.Println("Listening on port", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}

func extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf(errMissingHeader)
	}

	// Extract the token from the Bearer string
	token := strings.TrimPrefix(authHeader, "Bearer ")
	return token, nil
}

func createItem(w http.ResponseWriter, r *http.Request) {
	token, err := extractToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	validToken := false
	for _, user := range users {
		if user.Token == token {
			validToken = true
			break
		}
	}

	if !validToken {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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

	// Find the highest ID among the existing items
	highestID := 0
	for _, item := range items {
		if item.ID > highestID {
			highestID = item.ID
		}
	}

	// Assign a unique ID to the new item
	mu.Lock()
	item.ID = highestID + 1
	store[item.ID] = item
	mu.Unlock()

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

func updateItems(w http.ResponseWriter, r *http.Request) {
	token, err := extractToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	validToken := false
	for _, user := range users {
		if user.Token == token {
			validToken = true
			break
		}
	}

	if !validToken {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	// Parse the request body
	var updateRequest struct {
		IDs  []int `json:"ids"`
		Item Item  `json:"item"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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

	// Update the items with the given IDs
	updatedIDs := make(map[int]bool)
	for _, id := range updateRequest.IDs {
		for i, item := range items {
			if item.ID == id {
				if updateRequest.Item.Name != "" {
					items[i].Name = updateRequest.Item.Name
				}
				if updateRequest.Item.Price != 0 {
					items[i].Price = updateRequest.Item.Price
				}
				updatedIDs[id] = true
			}
		}
	}

	// Write the items back to the file
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

	// Respond with the updated IDs
	json.NewEncoder(w).Encode(updatedIDs)
}

func decodeRequestBody(w http.ResponseWriter, r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(v)
	if err != nil {
		http.Error(w, errBadRequest, http.StatusBadRequest)
	}
	return err
}

func login(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := decodeRequestBody(w, r, &user); err != nil {
		return
	}

	for _, u := range users {
		if u.Email == user.Email {
			err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(user.Password))
			if err != nil {
				http.Error(w, "Invalid password", http.StatusUnauthorized)
				return
			}

			json.NewEncoder(w).Encode(Token{Token: u.Token})
			return
		}
	}

	http.Error(w, "User not found", http.StatusNotFound)
}

func register(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := decodeRequestBody(w, r, &user); err != nil {
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user.ID = len(users) + 1
	user.Password = string(hashedPassword)
	user.Token = fmt.Sprintf("token%d", user.ID)

	mu.Lock()
	users[user.ID] = user
	mu.Unlock()

	// Create a new struct with only the fields you want to return
	responseUser := struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	}{ID: user.ID, Email: user.Email}

	json.NewEncoder(w).Encode(responseUser)
}
