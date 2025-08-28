package main

// Importing packages
import (
	"bufio"
	sql "database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-sql-driver/mysql"
)

// Insults represents an insult with an ID and the insult text
type Insults struct {
	ID     int64
	Insult string
}

// Jokes represents a joke with an ID, answer, and question
type Jokes struct {
	ID       int64
	Answer   string
	Question string
}

// db is a pointer to the sql database
var db *sql.DB

// main connects to the database and runs various functions to demonstrate functionality
func main() {

	// Connecting to the sql database
	cfg := mysql.NewConfig()
	cfg.User = os.Getenv("DBUSER")
	cfg.Passwd = os.Getenv("DBPASS")
	// Check if DBUSER and DBPASS environment variables are set
	if cfg.User == "" || cfg.Passwd == "" {
		log.Fatal("Environment variables DBUSER and DBPASS must be set")
	}

	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:3306"
	cfg.DBName = "Carnac"

	// Get a database handle
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

	// Return all insults from the current database
	insults, err := getInsults(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Insult: %v\n", insults)

	// Return all jokes from the current database
	jokes, err := getJokes(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Jokes: %v\n", jokes)

	// Find a joke by its ID
	var id int64
	fmt.Print("Enter the ID of the joke:")
	fmt.Scanln(&id)
	joke, err := getJokeById(id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Joke: %v\n", joke)

	// Find an insult by its ID
	var insultID int64
	fmt.Print("Enter the ID of the insult:")
	fmt.Scanln(&insultID)
	insult, err := getInsultsById(insultID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Insult: %v\n", insult)

	// Ask the user to add a new joke or insult to the database
	for {
		fmt.Println("Do you want to add a new joke or insult to the database? (j/i/n)")
		var choice string
		if _, err := fmt.Scanln(&choice); err != nil {
			fmt.Println("Invalid input, please enter either (j,i or n.")
			continue
		}
		switch choice {
		case "j":
			addJoke()
		case "i":
			addInsult()
		case "n":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid input, please enter either (j,i or n.")
			continue
		}
	}
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

// getInsultsById returns an insult from an sql database by its id
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

// addNewJoke prompts the user to add a new joke to the database
func addNewJoke() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter the joke answer first: ")
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(answer)

	fmt.Print("Enter the joke question: ")
	question, _ := reader.ReadString('\n')
	question = strings.TrimSpace(question)

	if question == "" || answer == "" {
		fmt.Println("Both the answer and question must be provided.")
		return
	}

	jokeID, err :== addJoke(Jokes{Answer: answer, Question: question})
	if err != nil {
		log.Fatalf("Error adding joke: %v", err)
	}
	fmt.Printf("Joke added with ID: %d\n", jokeID)
}

// addNewInsult prompts the user to add a new insult to the database
func addNewInsult() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter the insult: ")
	insult, _ := reader.ReadString('\n')
	insult = strings.TrimSpace(insult)

	if insult == "" {
		fmt.Println("Insult must be provided.")
		return
	}

	insultID, err := addInsult(Insults{Insult: insult})
	if err != nil {
		log.Fatalf("Error adding insult: %v", err)
	}
	fmt.Printf("Insult added with ID: %d\n", insultID)
}

// addInsult adds an insult to the sql database
func addInsult(ins Insults) (int64, error) {
	result, err := db.Exec("INSERT INTO Insults (insult) values (?)", ins.Insult)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, nil
	}
	return id, nil
}

// addJoke adds a joke to the sql database
func addJoke(jok Jokes) (int64, error) {
	result, err := db.Exec("INSERT INTO Jokes (answer, question) values (?, ?)", jok.Answer, jok.Question)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, nil
	}
	return id, nil
}
