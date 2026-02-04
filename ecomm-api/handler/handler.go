package handler

import (
	authDto "ecomm/ecomm-api/handler/dto/auth" // Добавить импорт для DTO аутентификации
	orderDto "ecomm/ecomm-api/handler/dto/order"
	productDto "ecomm/ecomm-api/handler/dto/product"
	"ecomm/ecomm-api/service"
	"ecomm/util"
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

	productRes, err := h.service.UpdateProduct(r.Context(), id, &updateProductReq)
	if err != nil {
		responseWithError(w, r, err)
		return
	}
	respondWithJSON(w, http.StatusOK, productRes)
}

func (h *handler) deleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := extractAndParseId(r)
	if err != nil {
		responseWithError(w, r, err)
		return
	}
	err = h.service.DeleteProduct(r.Context(), id)
	if err != nil {
		responseWithError(w, r, err)
		return
	}
	respondWithJSON(w, http.StatusNoContent, nil)
}

func (h *handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var createOrderReq orderDto.CreateOrderReq
	if err := json.NewDecoder(r.Body).Decode(&createOrderReq); err != nil {
		responseWithError(w, r, err)
		return
	}
	orderRes, err := h.service.CreateOrder(r.Context(), &createOrderReq)
	if err != nil {
		responseWithError(w, r, err)
		return
	}
	respondWithJSON(w, http.StatusCreated, orderRes)
}

// LoginUser обрабатывает запросы на вход пользователя.
func (h *handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req authDto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responseWithError(w, r, errors.New("invalid request payload"))
		return
	}

	// TODO: Здесь должна быть реальная логика проверки учетных данных пользователя (из БД, например)
	// и получение user ID.
	// Для примера, используем хардкод:
	if req.Email != "test@example.com" || req.Password != "password" {
		responseWithError(w, r, errors.New("invalid credentials"))
		return
	}
	userID := "some-user-id-from-db" // Замените на реальный userID после аутентификации

	// Генерация Access Token
	accessToken, err := service.GenerateToken(userID, util.GetAccessTokenExpiration())
	if err != nil {
		responseWithError(w, r, fmt.Errorf("failed to generate access token: %w", err))
		return
	}

	resp := authDto.LoginResponse{
		AccessToken: accessToken,
	}

	respondWithJSON(w, http.StatusOK, resp)
}
