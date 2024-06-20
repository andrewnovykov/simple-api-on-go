package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type Token struct {
	Token string `json:"token"`
}

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

type ResponseUser struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

const (
	errBadRequest     = "bad request"
	errUnauthorized   = "unauthorized"
	errNotFound       = "not found"
	errInternalServer = "internal server error"
	errMissingHeader  = "missing authorization header"
)

func decodeRequestBody(w http.ResponseWriter, r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(v)
	if err != nil {
		http.Error(w, errBadRequest, http.StatusBadRequest)
	}
	return err
}

func Login(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := decodeRequestBody(w, r, &user); err != nil {
		return
	}

	// Read users from the JSON file
	file, err := os.Open("users.json")
	if err != nil {
		http.Error(w, "Failed to open users file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	var users []User
	err = json.NewDecoder(file).Decode(&users)
	if err != nil {
		http.Error(w, "Failed to decode users file", http.StatusInternalServerError)
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

func Register(w http.ResponseWriter, r *http.Request) {
	var mu sync.RWMutex
	var user User
	if err := decodeRequestBody(w, r, &user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Read existing users from the file
	file, err := os.Open("users.json")
	if err != nil {
		http.Error(w, "Failed to open users file", http.StatusInternalServerError)
		return
	}

	var existingUsers []User
	err = json.NewDecoder(file).Decode(&existingUsers)
	file.Close()
	if err != nil && err != io.EOF {
		http.Error(w, "Failed to read users from file", http.StatusInternalServerError)
		return
	}

	// Check if a user with the same email already exists
	mu.Lock()
	for _, u := range existingUsers {
		if u.Email == user.Email {
			http.Error(w, "User with this email already exists", http.StatusBadRequest)
			mu.Unlock()
			return
		}
	}
	mu.Unlock()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user.ID = len(existingUsers) + 1
	user.Password = string(hashedPassword)
	user.Token = string(hashedPassword)

	// Append the new user
	existingUsers = append(existingUsers, user)

	// Write the users back to the file
	file, err = os.OpenFile("users.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		http.Error(w, "Failed to open users file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(existingUsers)
	if err != nil {
		http.Error(w, "Failed to write users to file", http.StatusInternalServerError)
		return
	}

	// Create a new struct with only the fields you want to return
	responseUser := ResponseUser{ID: user.ID, Email: user.Email}

	json.NewEncoder(w).Encode(responseUser)
}
