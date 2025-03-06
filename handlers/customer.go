package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/egor-markin/wallcraft-go-test-task/config"
	"github.com/egor-markin/wallcraft-go-test-task/database"
	"github.com/egor-markin/wallcraft-go-test-task/utils"
	"github.com/lib/pq"
)

type CustomerQueries interface {
	ListCustomers(ctx context.Context) ([]database.Customer, error)
	CreateCustomer(ctx context.Context, params database.CreateCustomerParams) (database.Customer, error)
	GetCustomer(ctx context.Context, id int32) (database.Customer, error)
	UpdateCustomer(ctx context.Context, params database.UpdateCustomerParams) (database.Customer, error)
	DeleteCustomer(ctx context.Context, id int32) (string, error)
}

type CustomerHandler struct {
	Queries CustomerQueries
}

type createCustomerRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
type updateCustomerRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
type customerResponse struct {
	ID        int32  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (h *CustomerHandler) CustomersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// GET /customers
		customers, err := h.Queries.ListCustomers(r.Context())
		if err != nil {
			writeInternalServerError(w, err)
			return
		}
		response := []customerResponse{}
		for _, customer := range customers {
			response = append(response, customerResponse{
				ID:        customer.ID,
				FirstName: customer.FirstName,
				LastName:  customer.LastName,
			})
		}
		writeServerResponse(w, http.StatusOK, response)
	case http.MethodPost:
		// POST /customers
		var customer createCustomerRequest
		if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
			writeServerParseError(w, err)
			return
		}

		if strings.TrimSpace(customer.FirstName) == "" {
			http.Error(w, "First name is required", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(customer.LastName) == "" {
			http.Error(w, "Last name is required", http.StatusBadRequest)
			return
		}

		createdCustomer, err := h.Queries.CreateCustomer(r.Context(), database.CreateCustomerParams{
			FirstName: customer.FirstName,
			LastName:  customer.LastName,
		})
		if err != nil {
			writeInternalServerError(w, err)
			return
		}
		writeServerResponse(w, http.StatusCreated, customerResponse{
			ID:        createdCustomer.ID,
			FirstName: createdCustomer.FirstName,
			LastName:  createdCustomer.LastName,
		})
	default:
		http.Error(w, config.MethodNotAllowedMsg, http.StatusMethodNotAllowed)
	}
}

func (h *CustomerHandler) CustomerHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the customer ID from the URL path
	id, err := utils.ExtractTrailingID(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// GET /customers/{id}
		customer, err := h.Queries.GetCustomer(r.Context(), int32(id))
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Customer not found", http.StatusNotFound)
			} else {
				writeInternalServerError(w, err)
			}
			return
		}
		writeServerResponse(w, http.StatusOK, customerResponse{
			ID:        customer.ID,
			FirstName: customer.FirstName,
			LastName:  customer.LastName,
		})
	case http.MethodPatch:
		// PATCH /customers/{id}
		var customer updateCustomerRequest
		if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
			writeServerParseError(w, err)
			return
		}

		if strings.TrimSpace(customer.FirstName) == "" {
			http.Error(w, "First name is required", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(customer.LastName) == "" {
			http.Error(w, "Last name is required", http.StatusBadRequest)
			return
		}

		updatedCustomer, err := h.Queries.UpdateCustomer(r.Context(), database.UpdateCustomerParams{
			ID:        int32(id),
			FirstName: customer.FirstName,
			LastName:  customer.LastName,
		})
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Customer not found", http.StatusNotFound)
			} else {
				writeInternalServerError(w, err)
			}
			return
		}
		writeServerResponse(w, http.StatusOK, customerResponse{
			ID:        updatedCustomer.ID,
			FirstName: updatedCustomer.FirstName,
			LastName:  updatedCustomer.LastName,
		})
	case http.MethodDelete:
		// DELETE /customers/{id}
		deletionResult, err := h.Queries.DeleteCustomer(r.Context(), int32(id))
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) {
				// Check if it's a foreign key violation
				if pqErr.Code == "23503" { // 23503 is the SQLSTATE code for foreign key violation
					// Check the constraint name
					if pqErr.Constraint == "invoice_customer_id_fkey" {
						http.Error(w, "cannot delete customer: customer is referenced in the invoice table", http.StatusConflict)
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
		if deletionResult == "customer_not_found" {
			http.Error(w, "Customer not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, config.MethodNotAllowedMsg, http.StatusMethodNotAllowed)
	}
}
