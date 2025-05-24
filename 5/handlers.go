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

func formHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		claims, err := validateJWT(r)
		tmplPath := filepath.Join("templates", "form.html")
		tmpl := template.Must(template.New("form.html").Funcs(template.FuncMap{
			"contains": contains,
		}).ParseFiles(tmplPath))

		data := struct {
			Application *Application
			Languages   []string
			LoggedIn    bool
		}{
			Languages: languages,
		}

		if err == nil {
			// Конвертируем string Login в int
			loginInt, err := strconv.Atoi(claims.Login)
			if err != nil {
				http.Error(w, "Invalid user ID", http.StatusInternalServerError)
				return
			}

			app, err := getApplicationByLogin(loginInt)
			if err == nil {
				data.Application = app
				data.LoggedIn = true
			}
		}

		tmpl.Execute(w, data)
	}
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()

		app := &Application{
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
			// Обновление существующей заявки
			loginInt, err := strconv.Atoi(claims.Login)
			if err != nil {
				http.Error(w, "Invalid user ID", http.StatusInternalServerError)
				return
			}
			app.ID = loginInt

			err = updateApplication(app)
			if err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Создание новой заявки
		login := generateRandomString(8)
		password := generateRandomString(12)

		err = saveApplication(app, login, password)
		if err != nil {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		data := struct {
			Login    string
			Password string
		}{
			Login:    login,
			Password: password,
		}

		tmplPath := filepath.Join("templates", "success.html")
		tmpl := template.Must(template.ParseFiles(tmplPath))
		tmpl.Execute(w, data)
	}
}

// Функция для проверки наличия элемента в slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// loginHandler обрабатывает вход пользователя
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

		// Проверяем существование пользователя
		var dbPassword string
		err := db.QueryRow("SELECT Password FROM User WHERE Login = ?", login).Scan(&dbPassword)
		if err != nil {
			http.Error(w, "Invalid login or password", http.StatusUnauthorized)
			return
		}

		// Сравниваем пароли
		if err := bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(password)); err != nil {
			http.Error(w, "Invalid login or password", http.StatusUnauthorized)
			return
		}

		// Генерируем JWT токен
		token, err := generateJWT(login)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		// Устанавливаем cookie
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

// logoutHandler обрабатывает выход пользователя
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

// updateHandler обрабатывает обновление данных
func updateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	app := &Application{
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
