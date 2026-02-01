package main

import (
	"ecomm/db"
	"log"
)

func main() {
	db, err := db.NewDatabase()
	if err != nil {
		log.Fatal("error opening database: %v", err)
	}
	defer db.Close()
	log.Println("successfully connected to database")
}
