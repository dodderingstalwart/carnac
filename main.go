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
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

// Insults represents an insult with an ID and the insult text
type Insults struct {
	ID     int64  `json:"id"`
	Insult string `json:"insult"`
}

// Jokes represents a joke with an ID, answer, and question
type Jokes struct {
	ID       int64  `json:"id"`
	Answer   string `json:"answer"`
	Question string `json:"question"`
}

type CarnacEntry struct {
	ID       int64  `json:"id"`
	Answer   string `json:"answer,omitempty"`
	Question string `json:"question,omitempty"`
	Insult   string `json:"insult,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// db is a pointer to the sql database
var db *sql.DB

// main connects to the database and runs various functions to demonstrate functionality
func main() {
	// Subcommand flag definitions
	cmdGetInsultById := flag.Int64("insult-id", 0, "Find an insult by ID")
	cmdGetJokeById := flag.Int64("joke-id", 0, "Find a joke by ID")
	cmdGetJokeList := flag.Bool("list-jokes", false, "List all jokes")
	cmdGetInsultList := flag.Bool("list-insults", false, "List all insults")
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
		cfg.DBName = "carnac"

		if cfg.User == "" || cfg.Passwd == "" {
			log.Fatal("DBUSER and DBPASS are not set")
		}
		if err := exportToJSON(cfg.FormatDSN(), *cmdExport); err != nil {
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
	r := mux.NewRouter()

	// Enable CORS for all routes
	r.Use(corsMiddleware)

	// Legacy endpoints
	r.HandleFunc("/joke", jokeHandler)
	r.HandleFunc("/insult", insultHandler)
	r.HandleFunc("/export", exportHandler)
	r.HandleFunc("/status", statusHandler)
	r.HandleFunc("/ready", readyHandler)

	// Rest API endpoints
	api := r.PathPrefix("/api").Subrouter()

	//Health check endpoints
	api.HandleFunc("/health", healthCheckHandler).Methods("GET", "OPTIONS")

	// Jokes endpoints
	api.HandleFunc("/jokes", listJokesHandler).Methods("GET", "OPTIONS")
	api.HandleFunc("/jokes", createJokeHandler).Methods("POST", "OPTIONS")
	api.HandleFunc("/jokes/random", randomJokeHandler).Methods("GET", "OPTIONS")
	api.HandleFunc("/jokes/{id:[0-9]+}", getJokeByIdHandler).Methods("GET", "OPTIONS")

	//Insults endpoints
	api.HandleFunc("/insults", listInsultsHandler).Methods("GET", "OPTIONS")
	api.HandleFunc("/insults", createInsultsHandler).Methods("POST", "OPTIONS")
	api.HandleFunc("/insults/random", randomInsultsHandler).Methods("GET", "OPTIONS")
	api.HandleFunc("/insults/{id:[0-9]+}", getInsultsByIdHandler).Methods("GET", "OPTIONS")

	fmt.Printf("Starting server on port %s...\n", port)
	fmt.Printf("Legacy endpoints: http://localhost:%s/joke, /insult, /export\n", port)
	fmt.Printf("API endpoints: http://localhost:%s/api/jokes, /api/insults\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}

// CORS middleware to allow cross-origin requests
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})

}

// =================
// REST API Handlers
// =================

// healthCheckHandler checks if the server is running
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if err := db.Ping(); err != nil {
		respondError(w, http.StatusServiceUnavailable, "Database not healthy")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{
		"status":   "healthy",
		"database": "connected",
	})
}

// listJokesHandler lists all jokes in the database
func listJokesHandler(w http.ResponseWriter, r *http.Request) {
	jokes, err := getJokes(db)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve jokes")
		return
	}
	respondJSON(w, http.StatusOK, jokes)
}

// randomJokeHandler returns a random joke from the database
func randomJokeHandler(w http.ResponseWriter, r *http.Request) {
	var joke Jokes
	err := db.QueryRow("SELECT id, answer, question FROM Jokes ORDER BY RAND() LIMIT 1").Scan(&joke.ID, &joke.Answer, &joke.Question)

	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "No jokes found")
		return
	}

	if err != nil {
		log.Printf("Error retrieving random joke: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to retrieve random joke")
		return
	}

	respondJSON(w, http.StatusOK, joke)
}

// getJokeByIdHandler retrieves a joke by its ID from the database
func getJokeByIdHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid joke ID")
		return
	}

	joke, err := getJokeById(id)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "Joke not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch joke")
		return
	}
	respondJSON(w, http.StatusOK, joke)
}

// createJokeHandler creates a new joke in the database
func createJokeHandler(w http.ResponseWriter, r *http.Request) {
	var joke Jokes
	if err := json.NewDecoder(r.Body).Decode(&joke); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if joke.Answer == "" || joke.Question == "" {
		respondError(w, http.StatusBadRequest, "Both answer and question are required")
		return
	}

	id, err := addJoke(joke)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create joke")
		return
	}

	joke.ID = id
	respondJSON(w, http.StatusCreated, joke)
}

// listInsultsHandler lists all insults in the database
func listInsultsHandler(w http.ResponseWriter, r *http.Request) {
	insults, err := getInsults(db)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch insults")
		return
	}
	respondJSON(w, http.StatusOK, insults)
}

// randomInsultsHandler returns a random insult from the database
func randomInsultsHandler(w http.ResponseWriter, r *http.Request) {
	var insult Insults
	err := db.QueryRow("SELECT id, insult FROM Insults ORDER BY RAND() LIMIT 1").Scan(&insult.ID, &insult.Insult)

	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "No insults found")
		return
	}

	if err != nil {
		log.Printf("Error retrieving random insult: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to fetch random insult")
		return
	}

	respondJSON(w, http.StatusOK, insult)
}

// getInsultsByIdHandler retrieves an insult by its ID from the database
func getInsultsByIdHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid insult ID")
		return
	}

	insult, err := getInsultsById(id)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "Insult not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch insult")
		return
	}

	respondJSON(w, http.StatusOK, insult)
}

// createInsultsHandler creates a new insult in the database
func createInsultsHandler(w http.ResponseWriter, r *http.Request) {
	var ins Insults
	if err := json.NewDecoder(r.Body).Decode(&ins); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if ins.Insult == "" {
		respondError(w, http.StatusBadRequest, "Insult text is required")
		return
	}

	id, err := addInsult(ins)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create insult")
		return
	}

	ins.ID = id
	respondJSON(w, http.StatusCreated, ins)
}

// respondJSON sends a JSON response with the given status code and data
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

// redpondError sends a JSON error response with the given status code and message
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}

// ---------------
// Legacy Handlers
// ---------------

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
		if idParam != "" {
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
		} else {
			// Return all jokes if no ID is provided
			jokes, err := getJokes(db)
			if err != nil {
				http.Error(w, "Could not retrieve jokes", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(jokes)
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

// =================
// Database Functions
// =================

// initDB initializes the database connection
/*func initDB() error {
	var err error

	// Check database type from environment variable
	dbType := os.Getenv("DBTYPE")

	switch dbType {
	case "sqlite":
		dbPath := os.Getenv("DBPATH")
		if dbPath == "" {
			dbPath = "/data/carnac.db" // Default path for SQLite
		}

		db, err = sql.Open("sqlite3", dbPath)
		if err != nil {
			return fmt.Errorf("failed to open SQLite database: %v", err)
		}

		// Create tables if they don't exist
		if err := createTables(); err != nil {
			return err
		}

	case "mysql":
		// MySQL connection configuration
		cfg := mysql.NewConfig()
		cfg.User = os.Getenv("DBUSER")
		cfg.Passwd = os.Getenv("DBPASS")

		if cfg.User == "" || cfg.Passwd == "" {
			return fmt.Errorf("DBUSER and DBPASS environment variables must be set for MySQL")
		}

		cfg.Net = "tcp"
		cfg.Addr = os.Getenv("DBHOST") // Use DBHOST environment variable for MySQL host
		if cfg.Addr == "" {
			cfg.Addr = "localhost:3306" // Default MySQL address
		}
		cfg.DBName = "carnac"

		db, err = sql.Open("mysql", cfg.FormatDSN())
		if err != nil {
			return fmt.Errorf("failed to open MySQL database: %v", err)
		}

		if err := db.Ping(); err != nil {
			return fmt.Errorf("failed to connect to MySQL database: %v", err)
		}

	default:
		return fmt.Errorf("unsupported DBTYPE: %s. Use 'sqlite' or 'mysql'", dbType)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("database connection failed: %v", err)
	}

	log.Printf("Successfully connected to %s database:", dbType)
	return nil
}*/

func initDB() error {
	var err error

	// Determine database path
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		// Check if running in Fly.io (volume mounted at /data)
		if _, err := os.Stat("/data"); err == nil {
			dbPath = "/data/carnac.db"
		} else {
			// Local development
			dbPath = "./carnac.db"
		}
	}

	log.Printf("Using SQLite database at: %s", dbPath)

	// Open SQLite database
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open SQLite database: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// Create tables if they don't exist
	if err := createTables(); err != nil {
		return err
	}

	log.Println("Successfully connected to SQLite database")
	return nil
}

// createTables creates the necessary tables in the database if they do not already exist
func createTables() error {
	// Create Jokes table

	_, err := db.Exec(`
	    CREATE TABLE IF NOT EXISTS Jokes (
		    id INTEGER PRIMARY KEY AUTOINCREMENT,
			answer TEXT NOT NULL,
			question TEXT NOT NULL
		)
	
		`)
	if err != nil {
		return fmt.Errorf("failed to create Jokes table: %v", err)
	}

	// Create Insults table
	_, err = db.Exec(`
	    CREATE TABLE IF NOT EXISTS Insults (
		    id INTEGER PRIMARY KEY AUTOINCREMENT,
			insult TEXT NOT NULL
		)
	
	`)

	if err != nil {
		return fmt.Errorf("failed to create Insults table: %v", err)
	}

	log.Println("Database tabled created/verified")
	return nil
}

// getUserChoice prompts the user to enter a choice and returns it as an integer
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
