package handler

import (
	"ecomm/ecomm-api/service"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
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
		errEmailAlreadyExists      *service.ErrEmailAlreadyExists
		errInvalidEmailOrPassword  *service.ErrInvalidEmailOrPassword
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
	case errors.As(err, &errEmailAlreadyExists):
		status = http.StatusConflict
		clientMessage = fmt.Sprintf("User with email %s already exists", errEmailAlreadyExists.ID)
	case errors.As(err, &errInvalidEmailOrPassword):
		status = http.StatusBadRequest
		clientMessage = fmt.Sprintf("Invalid email or password")

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
