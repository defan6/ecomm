package handler

import (
	"ecomm/util"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
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

	r.Use(middleware.Recoverer)
	tokenValidator := util.NewJwtTokenValidator()
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(tokenValidator))
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
	})

	r.Post("/login", rm.AuthHandler.Authenticate)
	r.Post("/register", rm.AuthHandler.Register)

	return r
}

func Start(addr string) error {
	return http.ListenAndServe(addr, r)
}
