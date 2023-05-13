all: test build

# runs unit tests
.PHONY: test
test:
	go test ./...

# builds the server binary
.PHONY: build
build:
	go build -o openmcs cmd/openmcs/main.go

# creates seed data for a SQLite3 database
.PHONY: seed-sqlite3
seed-sqlite3:
	@for file in seed/sqlite3/*.sql; do \
  		echo $$file; \
  		cat $$file | sqlite3 data/game.db; \
  	done

# removes transient data including default databases
.PHONY: clean
clean:
	rm -f data/game.db
	rm -f openmcs
