package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

type (
	Application struct {
		Login     string
		FullName  string
		Phone     string
		Email     string
		Birthdate string
		Gender    string
		ProgLang  []string
		Bio       string
	}

	Statistics struct {
		Quantity   int
		ProgLang   map[string]int
		MostWanted string
	}

	AdminResponse struct {
		Applications []Application
		Statistics   Statistics
	}
)

func (appl Application) PLString() string {
	str := ""

	for _, pl := range appl.ProgLang {
		str += pl + ", "
	}

	return str[:len(str)-2]
}

func getPL(id string) ([]string, error) {
	pls := make([]string, 0)

	db, err := sql.Open("mysql", "u68874:3632703@/u68874")

	if err != nil {
		return nil, err
	}

	defer db.Close()

	sel, err := db.Query(`
		SELECT Name
		FROM PL_Application fav
		JOIN ProgLang pl
		ON pl.ProgLangID = fav.ProgLangID
		WHERE ApplicationID = ?;
	`, id)

	if err != nil {
		return nil, err
	}

	defer sel.Close()

	for sel.Next() {
		pl := ""

		err := sel.Scan(&pl)

		if err != nil {
			return nil, err
		}

		pls = append(pls, pl)
	}

	return pls, nil
}

func getApplications() ([]Application, error) {
	appls := make([]Application, 0)

	db, err := sql.Open("mysql", "u68874:3632703@/u68874")

	if err != nil {
		return nil, err
	}

	defer db.Close()

	sel, err := db.Query(`
		SELECT * FROM Application
	`)

	if err != nil {
		return nil, err
	}

	defer sel.Close()

	for sel.Next() {
		appl := Application{}

		err := sel.Scan(&appl.Login, &appl.FullName, &appl.Phone, &appl.Email, &appl.Birthdate, &appl.Gender, &appl.Bio)

		if err != nil {
			return nil, err
		}

		appl.ProgLang, err = getPL(appl.Login)

		if err != nil {
			return nil, err
		}

		appls = append(appls, appl)
	}

	return appls, nil
}

func getStatistics() (Statistics, error) {
	statistics := Statistics{
		ProgLang: make(map[string]int),
	}

	db, err := sql.Open("mysql", "u68874:3632703@/u68874")

	if err != nil {
		return statistics, err
	}

	defer db.Close()

	sel, err := db.Query(`
		SELECT Name, COUNT(*)
		FROM ProgLang
		JOIN PL_Application fav ON fav.ProgLangID = ProgLang.ProgLangID
		GROUP BY Name;
	`)

	if err != nil {
		return statistics, err
	}

	defer sel.Close()

	for sel.Next() {
		pl, count := "", 0

		err := sel.Scan(&pl, &count)

		if err != nil {
			return statistics, err
		}

		statistics.ProgLang[pl] = count
	}

	statistics.MostWanted = "c"

	for key, val := range statistics.ProgLang {
		if val > statistics.ProgLang[statistics.MostWanted] {
			statistics.MostWanted = key
		}
	}

	return statistics, nil
}

func administerHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("admin/administer.html")

	if err != nil {
		fmt.Fprintf(w, "Template error: %v", err)
		return
	}

	response := AdminResponse{}

	if r.Method == http.MethodPost {
		applications, err := getApplications()

		if err != nil {
			fmt.Fprintf(w, "MySQL error: %v", err)
			return
		}

		statistics, err := getStatistics()

		if err != nil {
			fmt.Fprintf(w, "MySQL error: %v", err)
			return
		}

		statistics.Quantity = len(applications)

		response.Applications = applications
		response.Statistics = statistics
	}

	tmpl.Execute(w, response)
}
