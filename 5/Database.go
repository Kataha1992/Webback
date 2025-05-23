package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type Application struct {
	ID        int
	FullName  string
	Phone     string
	Email     string
	Birthdate string
	Gender    string
	Biography string
	Languages []string
}

var db *sql.DB

func initDB() error {
	var err error
	db, err = sql.Open("mysql", "u68874:3632703@/u68874")
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	return nil
}

func generateRandomString(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal("Failed to generate random string:", err)
	}
	return base64.URLEncoding.EncodeToString(b)[:length]
}

func saveApplication(app *Application, login, password string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Хеширование пароля
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Сначала сохраняем заявку
	res, err := tx.Exec(`
		INSERT INTO Application 
		(FullName, PhoneNumber, Email, Birthdate, Gender, Biography) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		app.FullName, app.Phone, app.Email, app.Birthdate, app.Gender, app.Biography)
	if err != nil {
		return err
	}

	appID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// Затем сохраняем пользователя с ID = appID
	_, err = tx.Exec("INSERT INTO User (Login, Password) VALUES (?, ?)", appID, string(hash))
	if err != nil {
		return err
	}

	// Сохраняем языки программирования
	for _, lang := range app.Languages {
		var langID int
		err = tx.QueryRow("SELECT ProglangID FROM Proglang WHERE Name = ?", lang).Scan(&langID)
		if err != nil {
			return err
		}

		_, err = tx.Exec("INSERT INTO PL_Application (ApplicationID, ProglangID) VALUES (?, ?)", appID, langID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func getApplicationByLogin(login int) (*Application, error) {
	var app Application

	err := db.QueryRow(`
		SELECT ApplicationID, FullName, PhoneNumber, Email, 
		       Birthdate, Gender, Biography 
		FROM Application 
		WHERE ApplicationID = ?`, login).Scan(
		&app.ID, &app.FullName, &app.Phone, &app.Email,
		&app.Birthdate, &app.Gender, &app.Biography)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(`
		SELECT p.Name 
		FROM Proglang p
		JOIN PL_Application pa ON p.ProglangID = pa.ProglangID
		WHERE pa.ApplicationID = ?`, app.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var lang string
		if err := rows.Scan(&lang); err != nil {
			return nil, err
		}
		app.Languages = append(app.Languages, lang)
	}

	return &app, nil
}

func updateApplication(app *Application) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Обновляем основную информацию
	_, err = tx.Exec(`
		UPDATE Application 
		SET FullName = ?, PhoneNumber = ?, Email = ?, 
		    Birthdate = ?, Gender = ?, Biography = ? 
		WHERE ApplicationID = ?`,
		app.FullName, app.Phone, app.Email,
		app.Birthdate, app.Gender, app.Biography, app.ID)
	if err != nil {
		return err
	}

	// Удаляем старые связи с языками
	_, err = tx.Exec("DELETE FROM PL_Application WHERE ApplicationID = ?", app.ID)
	if err != nil {
		return err
	}

	// Добавляем новые связи с языками
	for _, lang := range app.Languages {
		var langID int
		err = tx.QueryRow("SELECT ProglangID FROM Proglang WHERE Name = ?", lang).Scan(&langID)
		if err != nil {
			return err
		}

		_, err = tx.Exec("INSERT INTO PL_Application (ApplicationID, ProglangID) VALUES (?, ?)", app.ID, langID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
