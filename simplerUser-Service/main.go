package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type Login struct {
	HashedPassword string
	SessionToken   string
	CSRFToken     string
}

var users = map[string]Login{}
func main() {

	http.HandleFunc("/register", register)

	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/protected", protected)
	http.ListenAndServe(":8080", nil)
}

//handlers
func register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost{
		er := http.StatusMethodNotAllowed
		http.Error(w, "Invalid method", er)
		return 
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	
	if len(username)<8 || len(password) <9 {
		er := http.StatusNotAcceptable
		http.Error(w, "Invalid username or passsword", er)
		return 
	}

	if _, ok := users[username]; ok {
		er := http.StatusConflict
		http.Error(w, "user already exists", er)
		return 
	}
	
	hashedPassword, _ := hashPassword(password)
	users[username] = Login{
		HashedPassword: hashedPassword,
	}
	fmt.Fprintln(w,"user registered successfully")
}

func login(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost {
		er := http.StatusMethodNotAllowed
		http.Error(w, "Invalid request method", er)
		return 
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	user, ok := users[username]
	if !ok || !checkHashedPassword(password, user.HashedPassword) {
		er := http.StatusUnauthorized
		http.Error(w, "Invalid username or password", er)
		return 
	}
	sessionToken := generateToken(6)
	csrfToken := generateToken(6)

	http.SetCookie(w, &http.Cookie{
		Name: "session_token",
		Value: sessionToken,
		Expires: time.Now().Add(2*time.Hour),
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name: "csrf_token",
		Value: csrfToken,
		Expires: time.Now().Add(2*time.Hour),
		HttpOnly: false,
	})

	user.SessionToken = sessionToken
	user.CSRFToken = csrfToken
	users[username] = user
	fmt.Fprintln(w,"login successfully")


}

func generateToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatal("Failed to genetate token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

func protected(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		err := http.StatusMethodNotAllowed
		http.Error(w, "Invlaid request method",err)
		return 
	}
	if err := Authorize(r); err != nil {
		er := http.StatusUnauthorized
		http.Error(w, "Unauthorized", er)
		return 
	}

	username := r.FormValue("username")
	fmt.Fprintf(w, "CSRF validation successfully, welcome, %s", username)
}

func logout(w http.ResponseWriter, r *http.Request) {
	if err := Authorize(r); err != nil {
		err := http.StatusUnauthorized
		http.Error(w, "Unathorized", err)
		return 
	}

	//clear cookies & csrf
	http.SetCookie(w, &http.Cookie{
		Name: "session_token",
		Value: "",
		Expires: time.Now().Add(-time.Hour),
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name: "csrf_token",
		Value: "",
		Expires: time.Now().Add(-time.Hour),
		HttpOnly: false,
	})

	//clear token from db
	username := r.FormValue("username")
	user, _ := users[username]
	user.SessionToken = ""
	user.CSRFToken = ""
	users[username] = user

	fmt.Fprintln(w, "user logged out successfully")

}