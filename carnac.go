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

var db *sql.DB

func main() {
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
	fmt.Println("Connected")

	insults, err := getInsults("May a near-sighted sand flea suck syrup off your short stack")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found: %v\n", insults)
}

func getInsults(joke string) ([]Insults, error) {
	var insults []Insults

	rows, err := db.Query("SELECT * FROM Insults", joke)
	if err != nil {
		return nil, fmt.Errorf("Insult %q: %v", joke, err)
	}
	defer rows.Close()

	for rows.Next() {
		var ins Insults
		if err := rows.Scan(&ins.ID, &ins.Insult); err != nil {
			return nil, fmt.Errorf("Insult %q: %v", joke, err)
		}
		insults = append(insults, ins)
	}
	return insults, nil
}
