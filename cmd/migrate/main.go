package main

import (
	"database/sql"
	"flag"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cmd := flag.String("cmd", "", "Command: up or down")
	flag.Parse()

	// 1. Open DB using the "sqlite" driver provided by the migration library (modernc)
	// We do not need to import glebarez here because 'database/sqlite' above
	// already registers a driver named "sqlite".
	db, err := sql.Open("sqlite", "dev.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 2. Create Migration Driver instance
	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// 3. Initialize Migrate
	m, err := migrate.NewWithDatabaseInstance(
		"file://sql/schema",
		"sqlite",
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}

	// 4. Run Command
	switch *cmd {
	case "up":
		err := m.Up()
		if err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		log.Println("Migrated UP successfully!")
	case "down":
		err := m.Down()
		if err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		log.Println("Migrated DOWN successfully!")
	default:
		log.Fatal("Unknown command. Use -cmd=up or -cmd=down")
	}
}
