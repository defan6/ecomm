package main

import (
	"ecomm/db"
	"ecomm/ecomm-api/handler"
	"ecomm/ecomm-api/service"
	"ecomm/ecomm-api/storer"
	"ecomm/util"
	"log"
)

func main() {
	db, err := db.NewDatabase()
	if err != nil {
		log.Fatal("error opening database: %v", err)
	}
	defer db.Close()
	log.Println("successfully connected to database")

	postgresStorer := storer.NewPostgresStorer(db.GetDB())
	passwordEncoder := util.NewPasswordEncoder()

	productService := service.NewProductService(postgresStorer)
	orderService := service.NewOrderService(postgresStorer)
	tokenGenerator := util.NewJwtTokenGenerator()
	authService := service.NewAuthService(postgresStorer, passwordEncoder, tokenGenerator)
	productHandler := handler.NewProductHandler(productService)
	orderHandler := handler.NewOrderHandler(orderService)
	authHandler := handler.NewAuthHandler(authService)
	routeManager := handler.NewRouterManager(authHandler, productHandler, orderHandler)
	routeManager.RegisterRoutes()
	handler.Start(":8080")
}
