package handler

import (
	"net/http"

	"github.com/go-chi/chi"
)

var r *chi.Mux

type RouterManager struct {
	AuthHandler    *AuthHandler
	ProductHandler *ProductHandler
	OrderHandler   *OrderHandler
}

func NewRouterManager(ah *AuthHandler, ph *ProductHandler, oh *OrderHandler) *RouterManager {
	return &RouterManager{
		AuthHandler:    ah,
		ProductHandler: ph,
		OrderHandler:   oh,
	}
}

func (rm *RouterManager) RegisterRoutes() *chi.Mux {
	r = chi.NewRouter()
	r.Route("/products", func(r chi.Router) {
		r.Post("/", rm.ProductHandler.createProduct)
		r.Get("/{id}", rm.ProductHandler.getProduct)
		r.Get("/", rm.ProductHandler.getProducts)
		r.Put("/{id}", rm.ProductHandler.updateProduct)
		r.Delete("/{id}", rm.ProductHandler.deleteProduct)
	})
	r.Route("/orders", func(r chi.Router) {
		r.Post("/", rm.OrderHandler.createOrder)
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/login", rm.AuthHandler.LoginUser)
		r.Post("/register", rm.AuthHandler.Register)
	})

	return r
}

func Start(addr string) error {
	return http.ListenAndServe(addr, r)
}
