package handler

import (
	"context"
	productDto "ecomm/ecomm-api/handler/dto/product"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

type ProductHandler struct {
	productService ProductService
}

type ProductService interface {
	CreateProduct(ctx context.Context, createProductReq *productDto.CreateProductReq) (productDto.ProductRes, error)
	GetProduct(ctx context.Context, id int64) (productDto.ProductRes, error)
	GetProducts(ctx context.Context) ([]productDto.ProductRes, error)
	UpdateProduct(ctx context.Context, id int64, updateProductReq *productDto.UpdateProductReq) (productDto.ProductRes, error)
	DeleteProduct(ctx context.Context, id int64) error
}

func NewProductHandler(productService ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
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
