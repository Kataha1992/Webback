package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	jwt "github.com/golang-jwt/jwt/v5"
)

type (
	Admin struct {
		Login    string
		Password string
	}
)

func valid(admin Admin) (bool, error) {
	db, err := sql.Open("mysql", "u68874:3632703@/u68874")

	if err != nil {
		return false, err
	}

	defer db.Close()

	login, password := "", ""

	sel, err := db.Query(`
		SELECT Login, Password 
		FROM ADMIN
	`)

	if err != nil {
		return false, err
	}

	defer sel.Close()

	for sel.Next() {
		err := sel.Scan(&login, &password)

		if err != nil {
			return false, err
		}
	}

	if login != admin.Login || password != fmt.Sprintf("%x", sha256.Sum256([]byte(admin.Password))) {
		return false, nil
	}

	return true, nil
}

func grantAccessToken(w http.ResponseWriter) {
	payload := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	key := []byte("access-token-secret-key")

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	t, err := accessToken.SignedString(key)

	if err != nil {
		fmt.Fprintf(w, "Ошибка при создании токена: %v", err)
		return
	}

	cookie := &http.Cookie{
		Name:  "accessToken",
		Value: t,
	}

	http.SetCookie(w, cookie)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("admin/login.html")

	if err != nil {
		fmt.Fprintf(w, "Template error: %v", err)
	}

	if r.Method == http.MethodPost {
		admin := Admin{
			Login:    r.FormValue("login"),
			Password: r.FormValue("password"),
		}

		ok, err := valid(admin)

		if err != nil {
			fmt.Fprintf(w, "MySQL error: %v", err)
			return
		}

		if !ok {
			tmpl.Execute(w, "Неверные логин или пароль")
			return
		}

		grantAccessToken(w)

		administerHandler(w, r)
		return
	}

	tmpl.Execute(w, nil)
}
