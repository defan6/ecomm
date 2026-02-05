package handler

import (
	"context"
	orderDto "ecomm/ecomm-api/handler/dto/order"
	"encoding/json"
	"net/http"
)

type OrderHandler struct {
	orderService OrderService
}

type OrderService interface {
	CreateOrder(ctx context.Context, createOrderReq *orderDto.CreateOrderReq) (orderDto.OrderRes, error)
}

func NewOrderHandler(orderService OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) createOrder(w http.ResponseWriter, r *http.Request) {
	var createOrderReq orderDto.CreateOrderReq
	if err := json.NewDecoder(r.Body).Decode(&createOrderReq); err != nil {
		responseWithError(w, r, err)
		return
	}
	userID, _ := GetUserIDFromContext(r.Context())
	createOrderReq.UserID = userID
	orderRes, err := h.orderService.CreateOrder(r.Context(), &createOrderReq)
	if err != nil {
		responseWithError(w, r, err)
		return
	}
	respondWithJSON(w, http.StatusCreated, orderRes)
}
