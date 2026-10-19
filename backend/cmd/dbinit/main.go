package main

import (
	"aiops-backend/database"
	"flag"
	"log"
	"os"
)

func main() {
	baseline := flag.Bool("baseline", false, "record the current schema version for an already initialized database without changing its tables or seed data")
	sqlOutput := flag.String("sql-output", "", "after initializing an empty build database, export its schema and initial data to this SQL file")
	flag.Parse()

	if *baseline {
		if *sqlOutput != "" {
			log.Fatal("--baseline and --sql-output cannot be used together")
		}
		if err := database.BaselineExistingDatabase(); err != nil {
			log.Fatal("baseline database: ", err)
		}
		log.Println("database baseline recorded; schema version", database.CurrentSchemaVersion)
		return
	}
	if *sqlOutput != "" {
		if _, err := os.Stat(*sqlOutput); err == nil {
			log.Fatal("initialization SQL already exists: ", *sqlOutput)
		} else if !os.IsNotExist(err) {
			log.Fatal("inspect initialization SQL path: ", err)
		}
	}

	if err := database.InitializeFreshDatabase(); err != nil {
		log.Fatal("initialize database: ", err)
	}
	if *sqlOutput != "" {
		file, err := os.OpenFile(*sqlOutput, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err != nil {
			log.Fatal("create initialization SQL: ", err)
		}
		if err := database.ExportInitializationSQL(file); err != nil {
			_ = file.Close()
			_ = os.Remove(*sqlOutput)
			log.Fatal("export initialization SQL: ", err)
		}
		if err := file.Close(); err != nil {
			log.Fatal("close initialization SQL: ", err)
		}
		log.Println("initialization SQL exported to", *sqlOutput)
	}
	log.Println("database initialized; schema version", database.CurrentSchemaVersion)
}
