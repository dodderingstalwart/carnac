package main

// Importing packages
import (
	"bufio"
	sql "database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
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

type CarnacEntry struct {
	ID       int64  `json:"id"`
	Answer   string `json:"answer,omitempty"`
	Question string `json:"question,omitempty"`
	Insult   string `json:"insult,omitempty"`
}

// db is a pointer to the sql database
var db *sql.DB

// main connects to the database and runs various functions to demonstrate functionality
func main() {
	// Subcommand flag definitions
	cmdGetInsultById := flag.Int64("insult-id", 0, "Find an insult by ID")
	cmdGetJokeById := flag.Int64("joke-id", 0, "Find a joke by ID")
	cmdGetJokeList := flag.Bool("List-jokes", false, "List all jokes")
	cmdGetInsultList := flag.Bool("List-insults", false, "List all insults")
	cmdExport := flag.String("export", "", "Export the database to a JSON file")
	cmdInteractive := flag.Bool("interactive", false, "Run in interactive mode")
	cmdServer := flag.Bool("server", false, "Run as a web server")
	cmdPort := flag.String("port", "8080", "Port to run the web server on")
	cmdDbHost := flag.String("dbhost", "localhost:3306", "Database host address")

	flag.Parse()

	// Initialize the database connection
	if err := initDB(); err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}

	// Defer closing the database connection until the main function exits
	defer db.Close()

	// Handle subcommands
	switch {
	case *cmdServer:
		startHTTPServer(*cmdPort)
	case *cmdGetJokeList:
		displayAllJokes()
	case *cmdGetInsultList:
		displayAllInsults()
	case *cmdGetJokeById > 0:
		joke, err := getJokeById(int64(*cmdGetJokeById))
		if err != nil {
			log.Fatalf("Could not retrieve joke: %v", err)
		}
		fmt.Printf("Joke ID: %d, Answer: %s, Question: %s\n", joke.ID, joke.Answer, joke.Question)
	case *cmdGetInsultById > 0:
		insult, err := getInsultsById(int64(*cmdGetInsultById))
		if err != nil {
			log.Fatalf("Could not retrieve insult: %v", err)
		}
		fmt.Printf("Insult ID: %d, Insult: %s\n", insult.ID, insult.Insult)
	case *cmdExport != "":
		cfg := mysql.NewConfig()
		cfg.User = os.Getenv("DBUSER")
		cfg.Passwd = os.Getenv("DBPASS")
		cfg.Net = "tcp"
		cfg.Addr = *cmdDbHost
		cfg.DBName = "Carnac"

		if cfg.User == "" || cfg.Passwd == "" {
			log.Fatal("DBUSER and DBPASS are not set")
		}
		if err := export_to_json(cfg.FormatDSN(), *cmdExport); err != nil {
			log.Fatalf("Error exporting database to JSON: %v", err)
		}
		fmt.Printf("Database successfully exported to %s\n", *cmdExport)
	case *cmdInteractive:
		interactiveMenu()
	default:
		fmt.Println("Carnac Database Management")
		fmt.Println("\nUsage:")
		flag.PrintDefaults()
	}
}

func interactiveMenu() {
	for {
		showMainMenu()
		choice := getUserChoice()

		switch choice {
		case 1:
			findJokeById()
		case 2:
			findInsultById()
		case 3:
			addNewJoke()
		case 4:
			addNewInsult()
		case 5:
			displayAllInsults()
		case 6:
			displayAllJokes()
		case 7:
			//export_to_json()
		case 8:
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid choice, please try again.")
			continue
		}
	}
}

// startHTTPServer starts an HTTP server with various endpoints
func startHTTPServer(port string) {
	http.HandleFunc("/joke", jokeHandler)
	http.HandleFunc("/insult", insultHandler)
	http.HandleFunc("/export", exportHandler)
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/ready", readyHandler)

	fmt.Printf("Starting server on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}

// readyHandler checks if the database connection is alive
func readyHandler(w http.ResponseWriter, r *http.Request) {
	if err := db.Ping(); err != nil {
		http.Error(w, "Database not ready", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ready"))
}

// statusHandler handles health check requests
func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// jokeHandler handles requests related to jokes
func jokeHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		idParam := r.URL.Query().Get("id")
		if idParam == "" {
			id, err := strconv.ParseInt(idParam, 10, 64)
			if err != nil {
				http.Error(w, "Invalid 'id' parameter", http.StatusBadRequest)
				return
			}
			joke, err := getJokeById(id)
			if err != nil {
				http.Error(w, "Joke not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(joke)
		}
	case "POST":
		var jok Jokes
		if err := json.NewDecoder(r.Body).Decode(&jok); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id, err := addJoke(jok)
		if err != nil {
			http.Error(w, "Could not add joke", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int64{"id": id})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleExportCommand handles the export command to export the database to a JSON file
func handleExportCommand(output string) {
	cfg := mysql.NewConfig()
	cfg.User = os.Getenv("DBUSER")
	cfg.Passwd = os.Getenv("DBPASS")

	if cfg.User == "" || cfg.Passwd == "" {
		log.Fatal("DBUSER and DBPASS are not set")
	}

	if err := export_to_json(cfg.FormatDSN(), output); err != nil {
		log.Fatalf("Error exporting database to JSON: %v", err)
	} else {
		fmt.Printf("Database successfully exported %s\n", output)
	}
}

// insultHandler handles HTTP requests for insults
func insultHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		idParam := r.URL.Query().Get("id")
		if idParam != "" {
			id, err := strconv.ParseInt(idParam, 10, 64)
			if err != nil {
				http.Error(w, "Invalid 'id' parameter", http.StatusBadRequest)
				return
			}
			insult, err := getInsultsById(id)
			if err != nil {
				http.Error(w, "Insult not found", http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(insult)
		} else {
			insults, err := getInsults(db)
			if err != nil {
				http.Error(w, "Could not retrieve insults", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(insults)
		}
	case "POST":
		var ins Insults
		if err := json.NewDecoder(r.Body).Decode(&ins); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id, err := addInsult(ins)
		if err != nil {
			http.Error(w, "Could not add insult", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int64{"id": id})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// exportHandler exports all jokes and insults to JSON format
func exportHandler(w http.ResponseWriter, r *http.Request) {
	var entries []CarnacEntry

	// Fetch jokes
	jokes, _ := getJokes(db)
	for _, j := range jokes {
		entries = append(entries, CarnacEntry{
			ID:       j.ID,
			Answer:   j.Answer,
			Question: j.Question,
		})
	}

	// Fetch insults
	insults, _ := getInsults(db)
	for _, i := range insults {
		entries = append(entries, CarnacEntry{
			ID:     i.ID,
			Insult: i.Insult,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

// initDB initializes the database connection
func initDB() error {
	// Connecting to the sql database
	cfg := mysql.NewConfig()
	cfg.User = os.Getenv("DBUSER")
	cfg.Passwd = os.Getenv("DBPASS")
	// Check if DBUSER and DBPASS environment variables are set
	if cfg.User == "" || cfg.Passwd == "" {
		log.Fatal("Environment variables DBUSER and DBPASS are not set")
	}

	// Configure the database connection (adjust as needed)
	cfg.Net = "tcp"
	cfg.DBName = "Carnac"

	// Get a database handle
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		return err
	}
	return nil
}

func getUserChoice() int {
	var choice int
	fmt.Print("Enter your choice (1-7): ")
	if _, err := fmt.Scanln(&choice); err != nil {
		bufio.NewReader(os.Stdin).ReadLine()
		return 0
	}
	return choice
}

// displayAllInsults retrieves and displays all insults from the database
func displayAllInsults() {
	fmt.Println("All the insults in the database:")
	insults, err := getInsults(db)
	if err != nil {
		log.Printf("Error could not retrieve insults: %v", err)
	} else {
		for _, ins := range insults {
			fmt.Printf("ID: %d, Insult: %s\n", ins.ID, ins.Insult)
		}
	}
}

// displayAllJokes retrieves and displays all jokes from the database
func displayAllJokes() {
	fmt.Println("All the jokes in the database:")
	jokes, err := getJokes(db)
	if err != nil {
		log.Printf("Error could not retrieve jokes: %v", err)
	} else {
		for _, jok := range jokes {
			fmt.Printf("ID: %d, Answer: %s, Question: %s\n", jok.ID, jok.Answer, jok.Question)
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

// findJokeById prompts the user to enter a joke ID and retrieves the corresponding joke from the database
func findJokeById() {
	var id int64
	fmt.Print("Enter the ID of the joke:")
	if _, err := fmt.Scanln(&id); err != nil {
		fmt.Println("Invalid input, please try again.")
		return
	}

	// Get the joke from the database
	joke, err := getJokeById(id)
	if err != nil {
		fmt.Println("No joke found with that ID.", err)
		return
	}
	fmt.Printf("The answer is: %s\n", joke.Answer)
	fmt.Printf("The question is: %s\n", joke.Question)
}

// findInsultById prompts the user to enter an insult ID and retrieves the corresponding insult from the database
func findInsultById() {
	var id int64
	fmt.Print("Enter the ID of the insult:")
	if _, err := fmt.Scanln(&id); err != nil {
		fmt.Println("Invalid input, please try again.")
		return
	}
	// Get the insult from the database
	insult, err := getInsultsById(id)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("No insult found with that ID.")
			return
		}
		log.Fatal(err)
	}
	fmt.Printf("Insult: %v\n", insult)
}

// addNewJoke prompts the user to add a new joke to the database
func addNewJoke() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter the joke with answer first: ")
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(answer)

	fmt.Print("Enter the joke question: ")
	question, _ := reader.ReadString('\n')
	question = strings.TrimSpace(question)

	if question == "" || answer == "" {
		fmt.Println("Both the answer and question must be provided.")
		return
	}

	jokeID, err := addJoke(Jokes{Answer: answer, Question: question})
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

func showMainMenu() {
	fmt.Println("\n***Carnac Main Menu***")
	fmt.Println("----------------")
	fmt.Println("1. Find a joke by ID")
	fmt.Println("2. Find an insult by ID")
	fmt.Println("3. Add a new joke")
	fmt.Println("4. Add a new insult")
	fmt.Println("5. Display all insults")
	fmt.Println("6. Display all jokes")
	fmt.Println("7. Export database to JSON file")
	fmt.Println("8. Exit")
}

func exportToJSON(dbString string, output string) error {
	// Connect to the database
	db, err := sql.Open("mysql", dbString)
	if err != nil {
		return fmt.Errorf("could not connect to the database: %v", err)
	}
	defer db.Close()

	var entries []CarnacEntry

	// Fetch jokes
	jokes, err := db.Query("SELECT answer, question FROM Jokes")
	if err != nil {
		return fmt.Errorf("could not fetch jokes: %v", err)
	}
	defer jokes.Close()

	for jokes.Next() {
		var entry CarnacEntry
		err = jokes.Scan(&entry.Answer, &entry.Question)
		if err != nil {
			return fmt.Errorf("could not scan joke: %v", err)
		}
		entries = append(entries, entry)
	}

	// Fetch insults
	insults, err := db.Query("SELECT insult FROM Insults")
	if err != nil {
		return fmt.Errorf("could not fetch insults: %v", err)
	}
	defer insults.Close()

	for insults.Next() {
		var entry CarnacEntry
		err = insults.Scan(&entry.Insult)
		if err != nil {
			return fmt.Errorf("could not scan insult: %v", err)
		}
		entries = append(entries, entry)
	}

	// Converting to JSON and writing to file
	file, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("could not create output file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(entries)
}
