![alt text](https://upload.wikimedia.org/wikipedia/en/1/10/Carnac_the_Magnificent.jpg)
# Carnac
An opensource Go application to input Carnac jokes and/or insults and store them into a database.<br>
* Creates an SQL database for jokes and insults
* Gets user input to add jokes and insults
* Displays all the jokes and insults to the user
* REST API to attach to my website

# Installation
## Prerequisites
* Go version 1.21 or higher
* Git
* SQL database like
    * Mariadb/MySQL 8.0 or higher
    * sqlite3
 
# Getting Started
1. Clone the repository and change into the directory:
   ```
   git clone https://github.com/dodderingstalwart/carnac.git
   cd carnac
   ```
2. Install the dependencies:
   ```
   go mod download
   ```
3. Create the database for the application:
   ```
   CREATE DATABASE carnac
   ```
4. Setup your env variables
   ```
   DBHOST="Your hostname"
   DBUSER="Your username"
   DBPASS="Your password"
   DBNAME="carnac"
   DBPORT=5432
   ```
5. Build and run the application
   ```
   go build -o carnac
   ./carnac
   ```
   or run without building
   ```
   go run main.go
   ```
# Usage
A list of commands:
```
./carnac --help
```
```
-export string
        Export the database to a JSON file
  -insult-id int
        Find an insult by ID
  -interactive
        Run in interactive mode
  -joke-id int
        Find a joke by ID
  -list-insults
        List all insults
  -list-jokes
        List all jokes
  -port string
        Port to run the web server on (default "8080")
  -server
        Run as a web server
```
# Schema
## Insults
```
CREATE TABLE Insults (ID int NOT NULL AUTO_INCREMENT,
                      insult varchar(255) NOT NULL,
                      PRIMARY KEY (ID)
);
```
## Jokes
```
CREATE TABLE Jokes (ID int NOT NULL AUTO_INCREMENT,
                    Answer varchar(255) NOT NULL,
                    Question varchar(255) NOT NULL,
                    PRIMARY KEY (ID)
);
```
# Roadmap
* Input the database to an LLVM in order to create modern Carnac jokes and/or insults.
* Added it to my website for random jokes/insults.
