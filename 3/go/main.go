package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/cgi"
	"regexp"
	"strings"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
)

type FormUser struct {
	fullName, phone, email, birthdate, gender string
	progLang []string
	bio string
}

func validationCheck(user FormUser) []string {
	var pattern string
	var re *regexp.Regexp

	var ans []string

	pattern = `^([А-ЯA-Z][а-яa-z]+ ){2}[А-ЯA-Z][а-яa-z]+$`
	re = regexp.MustCompile(pattern)

	if !re.MatchString(user.fullName) {
		ans = append(ans, "Поле ФИО некорректно заполнено")
	}

	pattern = `^(\+7|8)\d{10}$`
	re = regexp.MustCompile(pattern)

	if !re.MatchString(user.phone) {
		ans = append(ans, "Поле Телефон некорректно заполнено")
	}

	pattern = `^[A-Za-z][\w\.]+@\w+\.[a-z]+$`
	re = regexp.MustCompile(pattern)

	if !re.MatchString(user.email) {
		ans = append(ans, "Поле Email некорректно заполнено")
	}

	if user.birthdate == "" {
		ans = append(ans, "Заполните поле Дата рождения")
	}
	
	if user.gender == "" {
		ans = append(ans, "Заполните поле Пол")
	}

	if len(user.progLang) == 0 {
		ans = append(ans, "Выберите из списка Любимый язык программирования")
	}

	if user.bio == "" {
		ans = append(ans, "Заполните поле Биография")
	}

	return ans
}

func addToDataBase(user FormUser, w http.ResponseWriter) {
	db, err := sql.Open("mysql", "u68874:3632703@/u68874")

	if err != nil {
		fmt.Fprintf(w, "Ошибка подключения: %v", err)
		return
	}

	defer db.Close()

	insert, err := db.Query(fmt.Sprintf("INSERT INTO Application (FullName, PhoneNumber, Email, Birthdate, Gender, Biography) VALUES ('%s', '%s', '%s', '%s', '%s', '%s')", user.fullName, user.phone, user.email, user.birthdate, user.gender, user.bio))

	if err != nil {
		fmt.Fprintf(w, "Ошибка добавления: %v", err)
		return
	}

	defer insert.Close()

	sel, err := db.Query("SELECT ApplicationID FROM Application ORDER BY ApplicationID DESC LIMIT 1")

	if err != nil {
		fmt.Fprintf(w, "Ошибка извлечения: %v", err)
		return
	}

	defer sel.Close()

	var id int
	for sel.Next() {
		err = sel.Scan(&id)
	}

	if err != nil {
		fmt.Fprintf(w, "Ошибка считывания: %v", err)
		return
	}

	for _, name := range user.progLang {
		sel, err := db.Query(fmt.Sprintf("SELECT ProgLangID FROM ProgLang WHERE Name='%s'", name))

		if err != nil {
			fmt.Fprintf(w, "Ошибка извлечения: %v", err)
			return
		}

		defer sel.Close()

		var plId int
		for sel.Next() {
			err = sel.Scan(&plId)
		}

		if err != nil {
			fmt.Fprintf(w, "Ошибка считывания: %v", err)
			return
		}

		insert, err := db.Query(fmt.Sprintf("INSERT INTO PL_Application (ApplicationID, ProgLangID) VALUES ('%d', '%d')", id, plId))

		if err != nil {
			fmt.Fprintf(w, "Ошибка добавления: %v", err)
			return
		}

		defer insert.Close()
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		fmt.Fprintf(w, "Произошла ошибка: %v", err)
		return
	}

	user := FormUser{r.FormValue("full_name"),
					 r.FormValue("phone"),
					 r.FormValue("email"),
					 r.FormValue("birthdate"),
					 r.FormValue("gender"),
					 r.PostForm["prog_lang[]"],
					 r.FormValue("bio")}

	res := validationCheck(user)

	funcMap := template.FuncMap{
		"contains": strings.Contains,
	}

	tmpl, err := template.New("../form.html").Funcs(funcMap).ParseFiles("../form.html")

	if err != nil {
		fmt.Fprintf(w, "Ошибка при работе с шаблоном: %v", err)
		return
	}

	tmpl.Execute(w, res)

	if len(res) == 0 {
		addToDataBase(user, w)
	}
}

func main() {
	cgi.Serve(http.HandlerFunc(postHandler))
}
