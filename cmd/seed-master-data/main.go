package main

import (
	"cellar-app/database"
	"log"
)

func main() {
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	if err := database.SeedMasterData(db); err != nil {
		log.Fatalf("Master data seed failed: %v", err)
	}
	log.Println("Master data seed completed")
}
