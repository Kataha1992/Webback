package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
)

func rmRowByID(id any, col string, table string) error {
	db, err := sql.Open("mysql", "u68874:3632703@/u68874")

	if err != nil {
		return err
	}

	defer db.Close()

	_, err = db.Exec(fmt.Sprintf(`
		DELETE FROM %s
		WHERE %s = ?
	`, table, col), id)

	if err != nil {
		return err
	}

	return nil
}

func removeUser(id string) error {
	intID, err := strconv.Atoi(id)

	if err != nil {
		return err
	}

	err = rmRowByID(intID, "ApplicationID", "Application")

	if err != nil {
		return err
	}

	login := "u" + strings.Repeat("0", 7-len(id)) + id

	err = rmRowByID(login, "Login", "User")

	if err != nil {
		return err
	}

	return nil
}

func removeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("login.html")

	if err != nil {
		fmt.Fprintf(w, "Template error: %v", err)
		return
	}

	if !isAuthorized(r) {
		response := LoginResponse{
			Message: "Для выполнения этого действия необходимо авторизоваться",
			Type:    "warning_red",
		}

		tmpl.Execute(w, response)
		return
	}

	id := r.URL.Query().Get("id")

	err = removeUser(id)

	if err != nil {
		fmt.Fprintf(w, "MySQL error: %v", err)
		return
	}

	response := LoginResponse{
		Message: "Пользователь успешно удален",
		Type:    "warning_green",
	}

	tmpl.Execute(w, response)
}
