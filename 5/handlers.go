package main

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var languages = []string{"Go", "JavaScript", "Python", "Java", "C++", "PHP", "Ruby"}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		claims, err := validateJWT(r)
		tmplPath := filepath.Join("templates", "form.html")
		tmpl := template.Must(template.New("form.html").Funcs(template.FuncMap{
			"contains": contains,
		}).ParseFiles(tmplPath))

		data := struct {
			FormData  *FormData
			Languages []string
			LoggedIn  bool
		}{
			Languages: languages,
		}

		if err == nil {
			loginInt, err := strconv.Atoi(claims.Login)
			if err == nil {
				formData, err := getFormDataByLogin(loginInt)
				if err == nil {
					data.FormData = formData
					data.LoggedIn = true
				}
			}
		}

		tmpl.Execute(w, data)
	}
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		app := &FormData{
			FullName:  r.FormValue("fullName"),
			Phone:     r.FormValue("phone"),
			Email:     r.FormValue("email"),
			Birthdate: r.FormValue("birthdate"),
			Gender:    r.FormValue("gender"),
			Biography: r.FormValue("biography"),
			Languages: r.Form["languages"],
		}

		claims, err := validateJWT(r)
		if err == nil {
			loginInt, err := strconv.Atoi(claims.Login)
			if err != nil {
				http.Error(w, "Invalid user ID", http.StatusInternalServerError)
				return
			}
			app.ID = loginInt

			if err := updateApplication(app); err != nil {
				http.Error(w, "Failed to update application: "+err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		login := generateRandomString(8)
		password := generateRandomString(12)

		if err := saveApplication(app, login, password); err != nil {
			http.Error(w, "Failed to save application: "+err.Error(), http.StatusInternalServerError)
			return
		}

		tmplPath := filepath.Join("templates", "success.html")
		tmpl := template.Must(template.ParseFiles(tmplPath))
		tmpl.Execute(w, struct {
			Login    string
			Password string
		}{
			Login:    login,
			Password: password,
		})
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmplPath := filepath.Join("templates", "login.html")
		tmpl := template.Must(template.ParseFiles(tmplPath))
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		login := r.FormValue("login")
		password := r.FormValue("password")

		var dbPassword string
		err := db.QueryRow("SELECT Password FROM User WHERE Login = ?", login).Scan(&dbPassword)
		if err != nil {
			http.Error(w, "Invalid login or password", http.StatusUnauthorized)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(password)); err != nil {
			http.Error(w, "Invalid login or password", http.StatusUnauthorized)
			return
		}

		token, err := generateJWT(login)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    token,
			Expires:  time.Now().Add(24 * time.Hour),
			Path:     "/",
			HttpOnly: true,
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		Path:     "/",
		HttpOnly: true,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		claims, err := validateJWT(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		loginInt, err := strconv.Atoi(claims.Login)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusInternalServerError)
			return
		}

		app := &FormData{
			ID:        loginInt,
			FullName:  r.FormValue("fullName"),
			Phone:     r.FormValue("phone"),
			Email:     r.FormValue("email"),
			Birthdate: r.FormValue("birthdate"),
			Gender:    r.FormValue("gender"),
			Biography: r.FormValue("biography"),
			Languages: r.Form["languages"],
		}

		if err := updateApplication(app); err != nil {
			http.Error(w, "Failed to update application: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
