package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
)

type Insults struct {
	ID     int64
	Insult string
}

type Jokes struct {
	ID       int64
	Answer   string
	Question string
}

var db *sql.DB

func main() {

	// Connecting to the sql database
	cfg := mysql.NewConfig()
	cfg.User = os.Getenv("DBUSER")
	cfg.Passwd = os.Getenv("DBPASS")
	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:3306"
	cfg.DBName = "Carnac"

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")

	insults, err := getInsults(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Insult: %v\n", insults)

	joke, err := getJokeById(1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Joke: %v\n", joke)

	insultID, err := addInsult(Insults{
		Insult: "May a weird city council man rezone your sister as a business district",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ID of added insult: %v\n", insultID)
}

// getInsults returns the insults from an sql database
func getInsults(db *sql.DB) ([]Insults, error) {
	var insults []Insults

	rows, err := db.Query("SELECT * FROM Insults")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ins Insults
		if err := rows.Scan(&ins.ID, &ins.Insult); err != nil {
			return nil, err
		}
		insults = append(insults, ins)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return insults, nil
}

func getJokeById(id int64) (Jokes, error) {
	var jok Jokes

	row := db.QueryRow("SELECT * FROM Jokes WHERE id = ?", id)
	if err := row.Scan(&jok.ID, &jok.Answer, &jok.Question); err != nil {
		if err == sql.ErrNoRows {
			return jok, err
		}
		return jok, err
	}
	return jok, nil
}

func addInsult(ins Insults) (int64, error) {
	result, err := db.Exec("INSERT INTO Carnac (insult) values (?)", ins.Insult)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, nil
	}
	return id, nil
}
