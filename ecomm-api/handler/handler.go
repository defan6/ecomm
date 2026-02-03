package handler

import (
	"ecomm/ecomm-api/handler/dto/product"
	"ecomm/ecomm-api/service"
	"ecomm/ecomm-api/storer"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
)

type APIErrorResponse struct {
	Error    string
	Status   int
	Endpoint string
	Method   string
	Time     time.Time
}

func responseWithError(w http.ResponseWriter, r *http.Request, err error) {

	var clientMessage = "Internal Server Error"
	var productNotFoundError *storer.ProductNotFoundError
	var apiError APIErrorResponse

	if errors.As(err, &productNotFoundError) {
		log.Printf("API Error: status=%d, details=%v, endpoint=%s", http.StatusNotFound, err, r.URL.Path)
		clientMessage = fmt.Sprintf("Product with id %d not found", productNotFoundError.ID)
		apiError = APIErrorResponse{
			Error:    clientMessage,
			Status:   http.StatusNotFound,
			Endpoint: r.URL.Path,
			Method:   r.Method,
			Time:     time.Now(),
		}
	} else {
		log.Printf("API Error: status=%d, details=%v, endpoint=%s", http.StatusInternalServerError, err, r.URL.Path)
		apiError = APIErrorResponse{
			Error:    clientMessage,
			Status:   http.StatusInternalServerError,
			Endpoint: r.URL.Path,
			Method:   r.Method,
			Time:     time.Now(),
		}
	}
	respondWithJSON(w, apiError.Status, apiError)
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

type handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) createProduct(w http.ResponseWriter, r *http.Request) {
	var createProductReq productDto.CreateProductReq
	if err := json.NewDecoder(r.Body).Decode(&createProductReq); err != nil {
		responseWithError(w, r, err)
		return
	}
	productRes, err := h.service.CreateProduct(r.Context(), &createProductReq)

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

func (h *handler) getProduct(w http.ResponseWriter, r *http.Request) {
	id, err := extractAndParseId(r)
	if err != nil {
		responseWithError(w, r, err)
		return
	}
	productRes, err := h.service.GetProduct(r.Context(), id)

	if err != nil {
		responseWithError(w, r, err)
		return
	}

	respondWithJSON(w, http.StatusOK, productRes)
}

func (h *handler) getProducts(w http.ResponseWriter, r *http.Request) {
	productRes, err := h.service.GetProducts(r.Context())
	if err != nil {
		responseWithError(w, r, err)
		return
	}
	respondWithJSON(w, http.StatusOK, productRes)
}

func (h *handler) updateProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		responseWithError(w, r, err)
		return
	}
	updateProductReq := productDto.UpdateProductReq{}
	if err := json.NewDecoder(r.Body).Decode(&updateProductReq); err != nil {
		responseWithError(w, r, err)
		return
	}

	productRes, err := h.service.UpdateProduct(r.Context(), i, &updateProductReq)
	if err != nil {
		responseWithError(w, r, err)
		return
	}
	respondWithJSON(w, http.StatusOK, productRes)
}

func (h *handler) deleteProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		responseWithError(w, r, err)
		return
	}
	err = h.service.DeleteProduct(r.Context(), i)
	if err != nil {
		responseWithError(w, r, err)
		return
	}
	respondWithJSON(w, http.StatusNoContent, nil)
}
