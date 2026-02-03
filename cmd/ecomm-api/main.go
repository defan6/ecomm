package main

import (
	"ecomm/db"
	"ecomm/ecomm-api/handler"
	"ecomm/ecomm-api/service"
	"ecomm/ecomm-api/storer"
	"log"
)

func main() {
	db, err := db.NewDatabase()
	if err != nil {
		log.Fatal("error opening database: %v", err)
	}
	defer db.Close()
	log.Println("successfully connected to database")

	postgres := storer.NewPostgresStorer(db.GetDB())
	srv := service.NewService(postgres)
	hdl := handler.NewHandler(srv)
	handler.RegisterRoutes(hdl)
	handler.Start(":8080")
}
