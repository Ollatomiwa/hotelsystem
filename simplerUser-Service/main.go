package main

import (
	"fmt"
	"net/http"
)

type Login struct {
	HashedPassword string
	SessionToken   string
	CSRFToken     string
}

var users = map[string]Login{}
func main() {

	http.HandleFunc("/register", register)

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
