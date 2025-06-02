package main

import (
	"net/http"
	"net/http/cgi"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	cgi.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		action := r.URL.Query().Get("action")

		if action == "administer" {
			administerHandler(w, r)
		} else {
			loginHandler(w, r)
		}
	}))
}
