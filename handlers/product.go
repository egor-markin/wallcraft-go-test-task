package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/egor-markin/wallcraft-go-test-task/config"
	"github.com/egor-markin/wallcraft-go-test-task/database"
	"github.com/egor-markin/wallcraft-go-test-task/utils"
	"github.com/lib/pq"
)

type ProductQueries interface {
	ListProducts(ctx context.Context) ([]database.Product, error)
	CreateProduct(ctx context.Context, params database.CreateProductParams) (database.Product, error)
	GetProduct(ctx context.Context, id int32) (database.Product, error)
	UpdateProduct(ctx context.Context, params database.UpdateProductParams) (database.Product, error)
	DeleteProduct(ctx context.Context, id int32) (string, error)
}

type ProductHandler struct {
	Queries ProductQueries
}

type createProductRequest struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	Price          string `json:"price"`
	AvailableItems int32  `json:"available_items"`
}
type updateProductRequest struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	Price          string `json:"price"`
	AvailableItems int32  `json:"available_items"`
}
type productResponse struct {
	ID             int32  `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Price          string `json:"price"`
	AvailableItems int32  `json:"available_items"`
}

func (h *ProductHandler) ProductsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// GET /products
		products, err := h.Queries.ListProducts(r.Context())
		if err != nil {
			writeInternalServerError(w, err)
			return
		}
		response := []productResponse{}
		for _, product := range products {
			response = append(response, productResponse{
				ID:             product.ID,
				Name:           product.Name,
				Description:    product.Description.String,
				Price:          product.Price,
				AvailableItems: product.AvailableItems,
			})
		}
		writeServerResponse(w, http.StatusOK, response)
	case http.MethodPost:
		// POST /products
		var product createProductRequest
		if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
			writeServerParseError(w, err)
			return
		}

		if strings.TrimSpace(product.Name) == "" {
			http.Error(w, "Product name is required", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(product.Price) == "" {
			http.Error(w, "Product price is required", http.StatusBadRequest)
			return
		}
		if _, err := strconv.ParseFloat(product.Price, 64); err != nil {
			http.Error(w, "Invalid price", http.StatusBadRequest)
			return
		}
		if product.AvailableItems < 0 {
			http.Error(w, "available_items must be greater than or equal to 0", http.StatusBadRequest)
			return
		}

		createdProduct, err := h.Queries.CreateProduct(r.Context(), database.CreateProductParams{
			Name:           product.Name,
			Description:    sql.NullString{String: product.Description, Valid: product.Description != ""},
			Price:          product.Price,
			AvailableItems: product.AvailableItems,
		})
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				if pqErr.Constraint == "product_available_items_check" {
					http.Error(w, "available_items must be greater than or equal to 0", http.StatusBadRequest)
				} else if pqErr.Constraint == "product_price_check" {
					http.Error(w, "price should be a positive number", http.StatusBadRequest)
				} else {
					writeInternalServerError(w, err)
				}
			} else {
				writeInternalServerError(w, err)
			}
			return
		}

		writeServerResponse(w, http.StatusCreated, productResponse{
			ID:             createdProduct.ID,
			Name:           createdProduct.Name,
			Description:    createdProduct.Description.String,
			Price:          createdProduct.Price,
			AvailableItems: createdProduct.AvailableItems,
		})
	default:
		http.Error(w, config.MethodNotAllowedMsg, http.StatusMethodNotAllowed)
	}
}

func (h *ProductHandler) ProductHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the product ID from the URL path
	id, err := utils.ExtractTrailingID(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// GET /products/{id}
		product, err := h.Queries.GetProduct(r.Context(), int32(id))
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Product not found", http.StatusNotFound)
			} else {
				writeInternalServerError(w, err)
			}
			return
		}
		writeServerResponse(w, http.StatusOK, productResponse{
			ID:             product.ID,
			Name:           product.Name,
			Description:    product.Description.String,
			Price:          product.Price,
			AvailableItems: product.AvailableItems,
		})
	case http.MethodPatch:
		// PATCH /products/{id}
		var product updateProductRequest
		if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
			writeServerParseError(w, err)
			return
		}

		if strings.TrimSpace(product.Name) == "" {
			http.Error(w, "Product name is required", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(product.Price) == "" {
			http.Error(w, "Product price is required", http.StatusBadRequest)
			return
		}
		if _, err := strconv.ParseFloat(product.Price, 64); err != nil {
			http.Error(w, "Invalid price", http.StatusBadRequest)
			return
		}
		if product.AvailableItems < 0 {
			http.Error(w, "available_items must be greater than or equal to 0", http.StatusBadRequest)
			return
		}

		updatedProduct, err := h.Queries.UpdateProduct(r.Context(), database.UpdateProductParams{
			ID:             int32(id),
			Name:           product.Name,
			Description:    sql.NullString{String: product.Description, Valid: product.Description != ""},
			Price:          product.Price,
			AvailableItems: product.AvailableItems,
		})
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Product not found", http.StatusNotFound)
			} else if pqErr, ok := err.(*pq.Error); ok {
				if pqErr.Constraint == "product_available_items_check" {
					http.Error(w, "available_items must be greater than or equal to 0", http.StatusBadRequest)
				} else {
					writeInternalServerError(w, err)
				}
			} else {
				writeInternalServerError(w, err)
			}
			return
		}

		writeServerResponse(w, http.StatusOK, productResponse{
			ID:             updatedProduct.ID,
			Name:           updatedProduct.Name,
			Description:    updatedProduct.Description.String,
			Price:          updatedProduct.Price,
			AvailableItems: updatedProduct.AvailableItems,
		})
	case http.MethodDelete:
		// DELETE /products/{id}
		deletionResult, err := h.Queries.DeleteProduct(r.Context(), int32(id))
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) {
				// Check if it's a foreign key violation
				if pqErr.Code == "23503" { // 23503 is the SQLSTATE code for foreign key violation
					// Check the constraint name
					if pqErr.Constraint == "invoice_item_product_id_fkey" {
						http.Error(w, "cannot delete product: product is referenced in the invoice_item table", http.StatusConflict)
					} else {
						writeInternalServerError(w, err)
					}
				} else {
					writeInternalServerError(w, err)
				}
			} else {
				writeInternalServerError(w, err)
			}
			return
		}
		if deletionResult == "product_not_found" {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, config.MethodNotAllowedMsg, http.StatusMethodNotAllowed)
	}
}
