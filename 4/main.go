package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/cgi"
	"regexp"
	"strings"
	"text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type FormUser struct {
	FullName  string
	Phone     string
	Email     string
	Birthdate string
	Gender    string
	ProgLang  []string
	Bio       string
	Errors    map[string]string
}

type FormData struct {
	User      FormUser
	Errors    []string
	HasErrors bool
}

func validationCheck(user *FormUser) map[string]string {
	errors := make(map[string]string)

	// Validate FullName (ФИО)
	pattern := `^([А-ЯЁA-Z][а-яёa-z]+ ){2}[А-ЯЁA-Z][а-яёa-z]+$`
	re := regexp.MustCompile(pattern)
	if !re.MatchString(user.FullName) {
		errors["full_name"] = "ФИО должно состоять из трёх слов, каждое с заглавной буквы (допустимы только буквы и пробелы)"
	}

	// Validate Phone
	pattern = `^(\+7|8)\d{10}$`
	re = regexp.MustCompile(pattern)
	if !re.MatchString(user.Phone) {
		errors["phone"] = "Телефон должен быть в формате +7XXXXXXXXXX или 8XXXXXXXXXX (11 цифр)"
	}

	// Validate Email
	pattern = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re = regexp.MustCompile(pattern)
	if !re.MatchString(user.Email) {
		errors["email"] = "Email должен быть в формате example@domain.com"
	}

	// Validate Birthdate
	if user.Birthdate == "" {
		errors["birthdate"] = "Поле Дата рождения обязательно для заполнения"
	}

	// Validate Gender
	if user.Gender == "" {
		errors["gender"] = "Поле Пол обязательно для заполнения"
	}

	// Validate Programming Languages
	if len(user.ProgLang) == 0 {
		errors["prog_lang"] = "Выберите хотя бы один язык программирования"
	}

	// Validate Biography
	if user.Bio == "" {
		errors["bio"] = "Поле Биография обязательно для заполнения"
	} else if len(user.Bio) > 500 {
		errors["bio"] = "Биография не должна превышать 500 символов"
	}

	return errors
}

func addToDataBase(user FormUser, w http.ResponseWriter) error {
	db, err := sql.Open("mysql", "u68874:3632703@/u68874")
	if err != nil {
		return fmt.Errorf("ошибка подключения: %v", err)
	}
	defer db.Close()

	insert, err := db.Query(fmt.Sprintf("INSERT INTO Application (FullName, PhoneNumber, Email, Birthdate, Gender, Biography) VALUES ('%s', '%s', '%s', '%s', '%s', '%s')",
		user.FullName, user.Phone, user.Email, user.Birthdate, user.Gender, user.Bio))
	if err != nil {
		return fmt.Errorf("ошибка добавления: %v", err)
	}
	defer insert.Close()

	sel, err := db.Query("SELECT ApplicationID FROM Application ORDER BY ApplicationID DESC LIMIT 1")
	if err != nil {
		return fmt.Errorf("ошибка извлечения: %v", err)
	}
	defer sel.Close()

	var id int
	for sel.Next() {
		err = sel.Scan(&id)
	}
	if err != nil {
		return fmt.Errorf("ошибка считывания: %v", err)
	}

	for _, name := range user.ProgLang {
		sel, err := db.Query(fmt.Sprintf("SELECT ProgLangID FROM ProgLang WHERE Name='%s'", name))
		if err != nil {
			return fmt.Errorf("ошибка извлечения: %v", err)
		}
		defer sel.Close()

		var plId int
		for sel.Next() {
			err = sel.Scan(&plId)
		}
		if err != nil {
			return fmt.Errorf("ошибка считывания: %v", err)
		}

		insert, err := db.Query(fmt.Sprintf("INSERT INTO PL_Application (ApplicationID, ProgLangID) VALUES ('%d', '%d')", id, plId))
		if err != nil {
			return fmt.Errorf("ошибка добавления: %v", err)
		}
		defer insert.Close()
	}

	return nil
}

func setFormCookies(w http.ResponseWriter, user FormUser) {
	// Set cookies for valid form data (1 year expiration)
	expiration := time.Now().Add(365 * 24 * time.Hour)
	http.SetCookie(w, &http.Cookie{Name: "full_name", Value: user.FullName, Expires: expiration})
	http.SetCookie(w, &http.Cookie{Name: "phone", Value: user.Phone, Expires: expiration})
	http.SetCookie(w, &http.Cookie{Name: "email", Value: user.Email, Expires: expiration})
	http.SetCookie(w, &http.Cookie{Name: "birthdate", Value: user.Birthdate, Expires: expiration})
	http.SetCookie(w, &http.Cookie{Name: "gender", Value: user.Gender, Expires: expiration})
	http.SetCookie(w, &http.Cookie{Name: "prog_lang", Value: strings.Join(user.ProgLang, ","), Expires: expiration})
	http.SetCookie(w, &http.Cookie{Name: "bio", Value: user.Bio, Expires: expiration})
}

func setErrorCookies(w http.ResponseWriter, errors map[string]string) {
	// Set error cookies (session only)
	for field, msg := range errors {
		http.SetCookie(w, &http.Cookie{Name: "error_" + field, Value: msg})
	}
}

func clearErrorCookies(w http.ResponseWriter) {
	// Clear all error cookies
	fields := []string{"full_name", "phone", "email", "birthdate", "gender", "prog_lang", "bio"}
	for _, field := range fields {
		http.SetCookie(w, &http.Cookie{Name: "error_" + field, Value: "", MaxAge: -1})
	}
}

func getFormFromCookies(r *http.Request) (FormUser, map[string]string) {
	user := FormUser{}
	errors := make(map[string]string)

	// Get form values from cookies
	if c, err := r.Cookie("full_name"); err == nil {
		user.FullName = c.Value
	}
	if c, err := r.Cookie("phone"); err == nil {
		user.Phone = c.Value
	}
	if c, err := r.Cookie("email"); err == nil {
		user.Email = c.Value
	}
	if c, err := r.Cookie("birthdate"); err == nil {
		user.Birthdate = c.Value
	}
	if c, err := r.Cookie("gender"); err == nil {
		user.Gender = c.Value
	}
	if c, err := r.Cookie("prog_lang"); err == nil {
		user.ProgLang = strings.Split(c.Value, ",")
	}
	if c, err := r.Cookie("bio"); err == nil {
		user.Bio = c.Value
	}

	// Get error messages from cookies
	for _, field := range []string{"full_name", "phone", "email", "birthdate", "gender", "prog_lang", "bio"} {
		if c, err := r.Cookie("error_" + field); err == nil {
			errors[field] = c.Value
		}
	}

	return user, errors
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Handle GET request - show form with data from cookies
		user, errors := getFormFromCookies(r)
		user.Errors = errors
		
		// Clear error cookies after displaying them
		clearErrorCookies(w)
		
		// Render form
		renderForm(w, FormData{User: user, HasErrors: len(errors) > 0})
		return
	}

	// Handle POST request
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Ошибка обработки формы", http.StatusBadRequest)
		return
	}

	user := FormUser{
		FullName:  r.FormValue("full_name"),
		Phone:     r.FormValue("phone"),
		Email:     r.FormValue("email"),
		Birthdate: r.FormValue("birthdate"),
		Gender:    r.FormValue("gender"),
		ProgLang:  r.PostForm["prog_lang[]"],
		Bio:       r.FormValue("bio"),
	}

	// Validate form
	errors := validationCheck(&user)
	if len(errors) > 0 {
		// Save errors and form data to cookies
		setErrorCookies(w, errors)
		for field, value := range map[string]string{
			"full_name":  user.FullName,
			"phone":      user.Phone,
			"email":      user.Email,
			"birthdate":  user.Birthdate,
			"gender":     user.Gender,
			"prog_lang":  strings.Join(user.ProgLang, ","),
			"bio":        user.Bio,
		} {
			http.SetCookie(w, &http.Cookie{Name: field, Value: value})
		}
		
		// Redirect back to form with GET
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
		return
	}

	// Form is valid - save to database
	if err := addToDataBase(user, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save valid form data to cookies for 1 year
	setFormCookies(w, user)
	
	// Clear any error cookies
	clearErrorCookies(w)
	
	// Show success message or redirect
	fmt.Fprintf(w, "Форма успешно отправлена!")
}

func renderForm(w http.ResponseWriter, data FormData) {
	funcMap := template.FuncMap{
		"contains": func(slice []string, item string) bool {
			for _, s := range slice {
				if s == item {
					return true
				}
			}
			return false
		},
	}

	tmpl, err := template.New("form.html").Funcs(funcMap).ParseFiles("form.html")
	if err != nil {
		http.Error(w, "Ошибка при работе с шаблоном: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка при отображении шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	cgi.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			if r.Method == "GET" {
				postHandler(w, r)
			} else if r.Method == "POST" {
				postHandler(w, r)
			} else {
				http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
			}
		} else {
			http.NotFound(w, r)
		}
	}))
}