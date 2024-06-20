package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/mux"
)

type Item struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

var store = make(map[int]Item)

func extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf(errMissingHeader)
	}

	// Extract the token from the Bearer string
	token := strings.TrimPrefix(authHeader, "Bearer ")
	return token, nil
}

func CreateItem(w http.ResponseWriter, r *http.Request) {
	var mu sync.RWMutex
	token, err := extractToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Read existing users from the file
	userfile, err := os.Open("users.json")
	if err != nil {
		http.Error(w, "Failed to open users file", http.StatusInternalServerError)
		return
	}

	var existingUsers []User
	err = json.NewDecoder(userfile).Decode(&existingUsers)
	userfile.Close()
	if err != nil && err != io.EOF {
		http.Error(w, "Failed to read users from file", http.StatusInternalServerError)
		return
	}

	validToken := false
	for _, user := range existingUsers {
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

func GetItems(w http.ResponseWriter, r *http.Request) {
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

func GetItem(w http.ResponseWriter, r *http.Request) {
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

func UpdateItems(w http.ResponseWriter, r *http.Request) {
	token, err := extractToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Read existing users from the file
	userfile, err := os.Open("users.json")
	if err != nil {
		http.Error(w, "Failed to open users file", http.StatusInternalServerError)
		return
	}

	var existingUsers []User
	err = json.NewDecoder(userfile).Decode(&existingUsers)
	userfile.Close()
	if err != nil && err != io.EOF {
		http.Error(w, "Failed to read users from file", http.StatusInternalServerError)
		return
	}

	validToken := false
	for _, user := range existingUsers {
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
