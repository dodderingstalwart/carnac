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

	// Test the connection to the database and output Connected! if successful
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

	jokes, err := getJokes(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Jokes: %v\n", jokes)

	var id int64
	fmt.Print("Enter the ID of the joke:")
	fmt.Scanln(&id)
	joke, err := getJokeById(id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Joke: %v\n", joke)

	var insultID int64
	fmt.Print("Enter the ID of the insult:")
	fmt.Scanln(&insultID)
	insult, err := getInsultsById(insultID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Insult: %v\n", insult)

	/*jokeID, err := addJoke(Jokes{
		Answer:   "Answer added ",
		Question: "Question added ",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ID of added joke: %v\n", jokeID)*/

	/*insultID, err := addInsult(Insults{
		Insult: "Insult added ",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ID of added insult: %v\n", insultID)*/
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

func getInsultsById(id int64) (Insults, error) {
	var ins Insults

	row := db.QueryRow("SELECT * FROM Insults WHERE id = ?", id)
	if err := row.Scan(&ins.ID, &ins.Insult); err != nil {
		if err == sql.ErrNoRows {
			return ins, err
		}
		return ins, err
	}
	return ins, nil
}

// getJokes returns the jokes from an sql database
func getJokes(db *sql.DB) ([]Jokes, error) {
	var jokes []Jokes

	rows, err := db.Query("SELECT * FROM Jokes")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var jok Jokes
		if err := rows.Scan(&jok.ID, &jok.Answer, &jok.Question); err != nil {
			return nil, err
		}
		jokes = append(jokes, jok)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return jokes, nil
}

// getJokeById returns a joke from an sql database by its id
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

/*func addInsult(ins Insults) (int64, error) {
	result, err := db.Exec("INSERT INTO Insults (insult) values (?)", ins.Insult)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, nil
	}
	return id, nil
}*/

/*func addJoke(jok Jokes) (int64, error) {
	result, err := db.Exec("INSERT INTO Jokes (answer, question) values (?, ?)", jok.Answer, jok.Question)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, nil
	}
	return id, nil
}*/
