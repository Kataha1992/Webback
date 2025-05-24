package main

import (
	"log"
	"net/http"
	"net/http/cgi"
)

func main() {
	// Инициализация базы данных
	err := initDB()
	if err != nil {
		log.Fatal("Database initialization failed:", err)
	}

	// Настройка маршрутов
	http.HandleFunc("/", formHandler)
	http.HandleFunc("/submit", submitHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/update", updateHandler)

	// Запуск CGI-сервера
	err = cgi.Serve(http.HandlerFunc(handler))
	if err != nil {
		log.Fatal("CGI error:", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Маршрутизация запросов
	switch r.URL.Path {
	case "/":
		formHandler(w, r)
	case "/submit":
		submitHandler(w, r)
	case "/login":
		loginHandler(w, r)
	case "/logout":
		logoutHandler(w, r)
	case "/update":
		updateHandler(w, r)
	default:
		http.NotFound(w, r)
	}
}
