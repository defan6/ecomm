package handler

import (
	"net/http"

	"github.com/go-chi/chi"
)

var r *chi.Mux

func RegisterRoutes(handler *handler) *chi.Mux {
	r = chi.NewRouter()
	r.Route("/products", func(r chi.Router) {
		r.Post("/", handler.createProduct)
		r.Get("/{id}", handler.getProduct)
		r.Get("/", handler.getProducts)
		r.Put("/{id}", handler.updateProduct)
		r.Delete("/{id}", handler.deleteProduct)
	})
	r.Route("/orders", func(r chi.Router) {
		r.Post("/", handler.createOrder)
	})

	return r
}

func Start(addr string) error {
	return http.ListenAndServe(addr, r)
}
