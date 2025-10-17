![alt text](https://upload.wikimedia.org/wikipedia/en/1/10/Carnac_the_Magnificent.jpg)
# Carnac
An opensource Go application to input Carnac jokes and/or insults and store them into a database.<br>
* Creates an SQL database for jokes and insults
* Gets user input to add jokes and insults
* Displays all the jokes and insults to the user

# Installation
## Prerequisites
* Go version 1.21 or higher
* Git
* SQL database like
    * Mariadb/MySQL 8.0 or higher
    * SQLite 3
    * PostgreSQL 12.0 or higher
 
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

# Roadmap
* Put the data into a bucket on a selected cloud provider.
* Input the database to an LLVM in order to create modern Carnac jokes and/or insults.
* That will create new Carnac jokes based on current events.
* Make it a command line tool that will be containerized.
