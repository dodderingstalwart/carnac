# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

Repository overview
- Language/runtime: Go (module: github.com/dodderingstalwart/carnac)
- Primary entrypoint: main.go (combines CLI, HTTP server, and DB access)
- Persistence: MySQL via github.com/go-sql-driver/mysql; tables Jokes and Insults
- Containerization: Dockerfile (multi-stage build)
- Schema helper: init.sql (DDL for MySQL)
- Docs: README.md

High-level architecture
- Execution modes
  - CLI flags: --server, --port, --interactive, --export <file>, --joke-id <n>, --insult-id <n>, --list-jokes, --list-insults, --dbhost <host:port>
  - Interactive TUI loop for adding/listing records when --interactive is set.
  - HTTP server when --server is set: endpoints /status, /ready, /joke, /insult, /export.
    - /status: liveness probe (200 OK)
    - /ready: readiness probe (pings DB)
    - /joke: GET by id query param; POST to add
    - /insult: GET by id (or list all), POST to add
    - /export: returns combined jokes+insults JSON
- Data model (MySQL)
  - Jokes: (ID, Answer, Question)
  - Insults: (ID, Insult)
- Configuration
  - Environment variables read by app: DBUSER, DBPASS (required)
  - Flags: --dbhost (defaults tcp to localhost:3306); DB name is hardcoded in code as "Carnac"
  - Note: Dockerfile sets DBPASSWORD, DBHOST, DBPORT, DBNAME which do not map 1:1 to app’s DBPASS/DB name usage; set DBUSER/DBPASS explicitly at runtime when using the container.
- File map (only the essentials)
  - main.go: all handlers, CLI flag parsing, DB init, CRUD helpers, export logic
  - init.sql: example schema
  - Dockerfile: multi-stage build for a minimal Alpine runtime
  - README.md: prerequisites and basic usage

Common commands
- Install deps
  ```bash path=null start=null
  go mod download
  ```
- Build
  ```bash path=null start=null
  go build -o carnac
  ```
- Run (CLI mode)
  ```bash path=null start=null
  DBUSER={{DBUSER}} DBPASS={{DBPASS}} ./carnac --list-jokes
  ```
- Run (interactive)
  ```bash path=null start=null
  DBUSER={{DBUSER}} DBPASS={{DBPASS}} ./carnac --interactive
  ```
- Run (HTTP server)
  ```bash path=null start=null
  DBUSER={{DBUSER}} DBPASS={{DBPASS}} ./carnac --server --port 8080
  # Health: curl -fsS http://localhost:8080/status
  ```
- Run without building
  ```bash path=null start=null
  DBUSER={{DBUSER}} DBPASS={{DBPASS}} go run main.go --server --port 8080
  ```
- Format and static checks (no extra linters configured)
  ```bash path=null start=null
  go fmt ./...
  go vet ./...
  ```
- Tests
  - No *_test.go files present. When tests are added:
    ```bash path=null start=null
    go test ./...
    go test -run TestName ./...
    ```
- Initialize database (MySQL)
  ```bash path=null start=null
  # Create DB (name expected by README is "carnac"; code uses "Carnac")
  mysql -h {{DBHOST}} -P {{DBPORT}} -u {{DBUSER}} -p -e 'CREATE DATABASE carnac;'
  # Apply schema
  mysql -h {{DBHOST}} -P {{DBPORT}} -u {{DBUSER}} -p carnac < init.sql
  ```
- Export data to JSON via CLI
  ```bash path=null start=null
  DBUSER={{DBUSER}} DBPASS={{DBPASS}} ./carnac --export out.json --dbhost {{DBHOST}}:{{DBPORT}}
  ```
- Docker
  ```bash path=null start=null
  docker build -t carnac:local .
  docker run --rm -p 8080:8080 \
    -e DBUSER={{DBUSER}} -e DBPASS={{DBPASS}} \
    carnac:local
  ```

Important notes from README.md
- Prereqs: Go 1.21+; a MySQL-compatible DB (MySQL/MariaDB). Clone repo, run go mod download, create the database, set DBUSER/DBPASS, build/run.
- CLI help: ./carnac --help lists all flags.

Warp-specific guidance
- Index the repo (small Go project). Prioritize reading main.go for behavior, Dockerfile for container usage, and init.sql for schema.
- When running commands that touch secrets, pass DB credentials via environment variables (DBUSER, DBPASS). Do not print them in the terminal.
