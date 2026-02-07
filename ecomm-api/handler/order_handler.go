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
	UpdateOrder(ctx context.Context, id int64, updateOrderRequest *orderDto.UpdateOrderReq) (orderDto.OrderRes, error)
	CancelOrder(ctx context.Context, id int64, currentUserId int64) (*orderDto.OrderRes, error)
	GetOrder(ctx context.Context, id int64) (*orderDto.OrderRes, error)
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

func (h *OrderHandler) getOrder(w http.ResponseWriter, r *http.Request) {
	id, err := extractAndParseId(r)
	if err != nil {
		responseWithError(w, r, err)
		return
	}
	res, err := h.orderService.GetOrder(r.Context(), id)
	if err != nil {
		responseWithError(w, r, err)
		return
	}
	respondWithJSON(w, http.StatusOK, res)
}

func (h *OrderHandler) updateOrder(w http.ResponseWriter, r *http.Request) {
	id, err := extractAndParseId(r)
	if err != nil {
		responseWithError(w, r, err)
		return
	}
	var updateOrderReq orderDto.UpdateOrderReq
	if err = json.NewDecoder(r.Body).Decode(&updateOrderReq); err != nil {
		responseWithError(w, r, err)
		return
	}
	userID, _ := GetUserIDFromContext(r.Context())
	updateOrderReq.UserID = userID
	orderRes, err := h.orderService.UpdateOrder(r.Context(), id, &updateOrderReq)
	if err != nil {
		responseWithError(w, r, err)
		return
	}
	respondWithJSON(w, http.StatusOK, orderRes)
}

func (h *OrderHandler) cancelOrder(w http.ResponseWriter, r *http.Request) {
	id, err := extractAndParseId(r)
	if err != nil {
		responseWithError(w, r, err)
		return
	}
	userID, _ := GetUserIDFromContext(r.Context())
	orderRes, err := h.orderService.CancelOrder(r.Context(), id, userID)
	if err != nil {
		responseWithError(w, r, err)
		return
	}
	respondWithJSON(w, http.StatusOK, orderRes)
}
