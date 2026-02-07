package handler

import (
	"ecomm/domain" // Added domain import
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
			// Routes requiring admin role
			r.Group(func(r chi.Router) {
				r.Use(AuthorizeMiddleware(domain.RoleAdmin))
				r.Post("/", rm.ProductHandler.createProduct)
				r.Put("/{id}", rm.ProductHandler.updateProduct)
				r.Delete("/{id}", rm.ProductHandler.deleteProduct)
			})
			// Routes accessible by any authenticated user (or no specific role check needed beyond authentication)
			r.Get("/{id}", rm.ProductHandler.getProduct)
			r.Get("/", rm.ProductHandler.getProducts)
		})
		r.Route("/orders", func(r chi.Router) {
			// Routes requiring user role
			r.Group(func(r chi.Router) {
				r.Use(AuthorizeMiddleware(domain.RoleUser))
				r.Get("/{id}", rm.OrderHandler.getOrder)
				r.Delete("/{id}", rm.OrderHandler.cancelOrder)
				r.Post("/", rm.OrderHandler.createOrder)
				r.Put("/{id}", rm.OrderHandler.updateOrder)
			})
		})
	})

	r.Post("/login", rm.AuthHandler.Authenticate)
	r.Post("/register", rm.AuthHandler.Register)

	return r
}

func Start(addr string) error {
	return http.ListenAndServe(addr, r)
}
