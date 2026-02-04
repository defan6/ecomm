package handler

import (
	"context"
	orderDto "ecomm/ecomm-api/handler/dto/order"
	productDto "ecomm/ecomm-api/handler/dto/product"
	"ecomm/ecomm-api/service"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
)

type ProductHandler struct {
	productService *service.ProductService
}

type ProductService interface {
	CreateOrder(ctx context.Context, createOrderReq *orderDto.CreateOrderReq) (orderDto.OrderRes, error)
	GetProduct(ctx context.Context, id int64) (productDto.ProductRes, error)
	GetProducts(ctx context.Context) ([]productDto.ProductRes, error)
	UpdateProduct(ctx context.Context, id int64, updateProductReq *productDto.UpdateProductReq) (productDto.ProductRes, error)
	DeleteProduct(ctx context.Context, id int64) error
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

type APIErrorResponse struct {
	Error    string
	Status   int
	Endpoint string
	Method   string
	Time     time.Time
}

func responseWithError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		clientMessage              = "Internal Server Error"
		errNotFound                *service.ErrNotFound
		errNotEnough               *service.ErrNotEnoughStock
		errNotFoundProductForOrder *service.ErrNotFoundProductForOrder
		apiError                   APIErrorResponse
		status                     = http.StatusInternalServerError
	)

	switch {
	case errors.As(err, &errNotFound):
		status = http.StatusNotFound
		clientMessage = fmt.Sprintf("Product with id %d not found", errNotFound.ID)

	case errors.As(err, &errNotEnough):
		status = http.StatusConflict
		clientMessage = fmt.Sprintf("not enough stock for product with id %d. Requested: %d, Available: %d",
			errNotEnough.ID, errNotEnough.Requested, errNotEnough.Available)
	case errors.As(err, &errNotFoundProductForOrder):
		status = http.StatusNotFound
		clientMessage = fmt.Sprintf("Some product for order not found")
	default:
		// оставляем Internal Server Error
	}

	log.Printf(
		"API Error: status=%d, details=%v, endpoint=%s",
		status, err, r.URL.Path,
	)

	apiError = APIErrorResponse{
		Error:    clientMessage,
		Status:   status,
		Endpoint: r.URL.Path,
		Method:   r.Method,
		Time:     time.Now(),
	}

	respondWithJSON(w, status, apiError)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{error: error marshalling response}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (h *ProductHandler) createProduct(w http.ResponseWriter, r *http.Request) {
	var createProductReq productDto.CreateProductReq
	if err := json.NewDecoder(r.Body).Decode(&createProductReq); err != nil {
		responseWithError(w, r, err)
		return
	}
	productRes, err := h.productService.CreateProduct(r.Context(), &createProductReq)

	if err != nil {
		responseWithError(w, r, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, productRes)
}

func extractAndParseId(r *http.Request) (int64, error) {
	id := chi.URLParam(r, "id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, errors.New("invalid id")
	}
	return i, nil
}

func (h *ProductHandler) getProduct(w http.ResponseWriter, r *http.Request) {
	id, err := extractAndParseId(r)
	if err != nil {
		responseWithError(w, r, err)
		return
	}
	productRes, err := h.productService.GetProduct(r.Context(), id)

	if err != nil {
		responseWithError(w, r, err)
		return
	}

	respondWithJSON(w, http.StatusOK, productRes)
}

func (h *ProductHandler) getProducts(w http.ResponseWriter, r *http.Request) {
	productRes, err := h.productService.GetProducts(r.Context())
	if err != nil {
		responseWithError(w, r, err)
		return
	}
	respondWithJSON(w, http.StatusOK, productRes)
}

func (h *ProductHandler) updateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := extractAndParseId(r)
	if err != nil {
		responseWithError(w, r, err)
		return
	}
	updateProductReq := productDto.UpdateProductReq{}
	if err := json.NewDecoder(r.Body).Decode(&updateProductReq); err != nil {
		responseWithError(w, r, err)
		return
	}

	productRes, err := h.productService.UpdateProduct(r.Context(), id, &updateProductReq)
	if err != nil {
		responseWithError(w, r, err)
		return
	}
	respondWithJSON(w, http.StatusOK, productRes)
}

func (h *ProductHandler) deleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := extractAndParseId(r)
	if err != nil {
		responseWithError(w, r, err)
		return
	}
	err = h.productService.DeleteProduct(r.Context(), id)
	if err != nil {
		responseWithError(w, r, err)
		return
	}
	respondWithJSON(w, http.StatusNoContent, nil)
}
