package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/egor-markin/wallcraft-go-test-task/config"
	"github.com/egor-markin/wallcraft-go-test-task/database"
	"github.com/lib/pq"
)

type InvoiceQueries interface {
	ListInvoices(ctx context.Context) ([]database.Invoice, error)
	CreateInvoice(ctx context.Context, params database.CreateInvoiceParams) (database.Invoice, error)
	GetInvoice(ctx context.Context, id int32) (database.Invoice, error)
	UpdateInvoice(ctx context.Context, params database.UpdateInvoiceParams) (database.UpdateInvoiceRow, error)
	DeleteInvoice(ctx context.Context, id int32) (string, error)
	ListProductsFromInvoice(ctx context.Context, invoiceID int32) ([]database.ListProductsFromInvoiceRow, error)
	AddProductToInvoice(ctx context.Context, params database.AddProductToInvoiceParams) (database.InvoiceItem, error)
	DeleteProductFromInvoice(ctx context.Context, params database.DeleteProductFromInvoiceParams) (string, error)
}

type InvoiceHandler struct {
	Queries InvoiceQueries
}

type createInvoiceRequest struct {
	InvoiceNumber string     `json:"invoice_number"`
	InvoiceDate   *time.Time `json:"invoice_date,omitempty"`
	CustomerID    int32      `json:"customer_id"`
}
type updateInvoiceRequest struct {
	InvoiceNumber string    `json:"invoice_number"`
	InvoiceDate   time.Time `json:"invoice_date"`
	CustomerID    int32     `json:"customer_id"`
}
type invoiceResponse struct {
	ID            int32     `json:"id"`
	InvoiceNumber string    `json:"invoice_number"`
	InvoiceDate   time.Time `json:"invoice_date"`
	CustomerID    int32     `json:"customer_id"`
}

type createInvoiceItemRequest struct {
	Count int32 `json:"count"`
}
type invoiceItemResponse struct {
	ID        int32 `json:"id"`
	InvoiceID int32 `json:"invoice_id"`
	ProductID int32 `json:"product_id"`
	Count     int32 `json:"count"`
}
type invoiceProductResponse struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Count       int32  `json:"count"`
	Sum         string `json:"sum"`
}

func (h *InvoiceHandler) InvoicesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// GET /invoices
		invoices, err := h.Queries.ListInvoices(r.Context())
		if err != nil {
			writeInternalServerError(w, err)
			return
		}
		response := []invoiceResponse{}
		for _, invoice := range invoices {
			response = append(response, invoiceResponse{
				ID:            invoice.ID,
				InvoiceNumber: invoice.InvoiceNumber,
				InvoiceDate:   invoice.InvoiceDate,
				CustomerID:    invoice.CustomerID,
			})
		}
		writeServerResponse(w, http.StatusOK, response)
	case http.MethodPost:
		// POST /invoices
		var invoiceCreate createInvoiceRequest
		if err := json.NewDecoder(r.Body).Decode(&invoiceCreate); err != nil {
			writeServerParseError(w, err)
			return
		}

		if strings.TrimSpace(invoiceCreate.InvoiceNumber) == "" {
			http.Error(w, "invoice_number must not be empty", http.StatusBadRequest)
			return
		}
		if invoiceCreate.CustomerID <= 0 {
			http.Error(w, "customer_id should be a positive number", http.StatusBadRequest)
			return
		}

		// invoiceDate is optional, if not provided, use the current time
		var invoiceDate time.Time
		if invoiceCreate.InvoiceDate != nil && !invoiceCreate.InvoiceDate.IsZero() {
			invoiceDate = *invoiceCreate.InvoiceDate
		} else {
			invoiceDate = time.Now()
		}

		createdInvoice, err := h.Queries.CreateInvoice(r.Context(), database.CreateInvoiceParams{
			InvoiceNumber: invoiceCreate.InvoiceNumber,
			InvoiceDate:   invoiceDate,
			CustomerID:    invoiceCreate.CustomerID,
		})
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) {
				switch pqErr.Code {
				case "23505":
					// Unique constraint violation
					http.Error(w, "Invoice number must be unique", http.StatusConflict)
					return
				case "23503":
					// Foreign key violation
					http.Error(w, "Specified customer does not exist", http.StatusBadRequest)
					return
				default:
					writeInternalServerError(w, err)
					return
				}
			}
			writeInternalServerError(w, err)
			return
		}

		writeServerResponse(w, http.StatusCreated, invoiceResponse{
			ID:            createdInvoice.ID,
			InvoiceNumber: createdInvoice.InvoiceNumber,
			InvoiceDate:   createdInvoice.InvoiceDate,
			CustomerID:    createdInvoice.CustomerID,
		})
	default:
		http.Error(w, config.MethodNotAllowedMsg, http.StatusMethodNotAllowed)
	}
}

func (h *InvoiceHandler) InvoiceHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Split the path into segments and filter out empty strings
	var segments []string
	for seg := range strings.SplitSeq(path, "/") {
		if seg != "" {
			segments = append(segments, seg)
		}
	}

	// Find the "invoices" segment
	invoiceIdx := -1
	for i, seg := range segments {
		if seg == "invoices" {
			invoiceIdx = i
			break
		}
	}
	if invoiceIdx == -1 || len(segments) <= invoiceIdx+1 {
		http.Error(w, "Invalid invoice path", http.StatusBadRequest)
		return
	}

	// Extract invoice ID
	invoiceID, err := strconv.Atoi(segments[invoiceIdx+1])
	if err != nil {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	// Check if there's a "products" segment after the invoice ID
	if len(segments) > invoiceIdx+2 && segments[invoiceIdx+2] == "products" {
		// Determine if a product ID is provided
		if len(segments) == invoiceIdx+3 {
			switch r.Method {
			case http.MethodGet:
				// GET /invoices/{invoice_id}/products
				items, err := h.Queries.ListProductsFromInvoice(r.Context(), int32(invoiceID))
				if err != nil {
					if err == sql.ErrNoRows {
						http.Error(w, "Invoice not found", http.StatusNotFound)
					} else {
						writeInternalServerError(w, err)
					}
					return
				}
				response := []invoiceProductResponse{}
				for _, item := range items {
					response = append(response, invoiceProductResponse{
						ID:          item.ID,
						Name:        item.Name,
						Description: item.Description.String,
						Price:       item.Price,
						Count:       item.Count,
						Sum:         item.Sum,
					})
				}
				writeServerResponse(w, http.StatusOK, response)
			default:
				http.Error(w, config.MethodNotAllowedMsg, http.StatusMethodNotAllowed)
			}
			return
		} else if len(segments) == invoiceIdx+4 {
			productID, err := strconv.Atoi(segments[invoiceIdx+3])
			if err != nil {
				http.Error(w, "Invalid product ID", http.StatusBadRequest)
				return
			}
			if r.Method == http.MethodDelete {
				// DELETE /invoices/{invoice_id}/products/{product_id}
				result, err := h.Queries.DeleteProductFromInvoice(r.Context(), database.DeleteProductFromInvoiceParams{InvoiceID: int32(invoiceID), ProductID: int32(productID)})
				if err != nil {
					writeInternalServerError(w, err)
					return
				}
				switch result {
				case "invoice_item_not_found":
					http.Error(w, "Provided invoice doesn't contain the specified product", http.StatusNotFound)
				default:
					w.WriteHeader(http.StatusNoContent)
				}
			} else if r.Method == http.MethodPost {
				// POST /invoices/{invoice_id}/products/{product_id}
				var params createInvoiceItemRequest
				if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
					writeServerParseError(w, err)
					return
				}

				if params.Count <= 0 {
					http.Error(w, "count must be greater than 0", http.StatusBadRequest)
					return
				}

				item, err := h.Queries.AddProductToInvoice(r.Context(), database.AddProductToInvoiceParams{
					InvoiceID: int32(invoiceID),
					ProductID: int32(productID),
					Count:     params.Count,
				})
				if err != nil {
					if pqErr, ok := err.(*pq.Error); ok {
						// Check if the error is a foreign key violation
						if pqErr.Code == "23503" { // 23503 is the SQLState code for foreign key violation
							constraint := pqErr.Constraint
							switch constraint {
							case "invoice_item_product_id_fkey":
								http.Error(w, "The provided product does not exist", http.StatusNotFound)
							case "invoice_item_invoice_id_fkey":
								http.Error(w, "The provided invoice does not exist", http.StatusNotFound)
							default:
								writeInternalServerError(w, err)
							}
						} else if pqErr, ok := err.(*pq.Error); ok {
							if pqErr.Constraint == "invoice_item_count_check" {
								http.Error(w, "count must be greater than 0", http.StatusBadRequest)
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
				writeServerResponse(w, http.StatusCreated, invoiceItemResponse{
					ID:        item.ID,
					InvoiceID: item.InvoiceID,
					ProductID: item.ProductID,
					Count:     item.Count,
				})
			} else {
				http.Error(w, config.MethodNotAllowedMsg, http.StatusMethodNotAllowed)
			}
			return
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
	}

	// Invoice-only endpoints: /invoices/{invoice_id}
	switch r.Method {
	case http.MethodGet:
		// GET /invoices/{invoice_id}
		invoice, err := h.Queries.GetInvoice(r.Context(), int32(invoiceID))
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Invoice not found", http.StatusNotFound)
			} else {
				writeInternalServerError(w, err)
			}
			return
		}
		writeServerResponse(w, http.StatusOK, invoiceResponse{
			ID:            invoice.ID,
			InvoiceNumber: invoice.InvoiceNumber,
			InvoiceDate:   invoice.InvoiceDate,
			CustomerID:    invoice.CustomerID,
		})
	case http.MethodPatch:
		// PATCH /invoices/{invoice_id}
		var invoiceUpdate updateInvoiceRequest
		if err := json.NewDecoder(r.Body).Decode(&invoiceUpdate); err != nil {
			writeServerParseError(w, err)
			return
		}

		if strings.TrimSpace(invoiceUpdate.InvoiceNumber) == "" {
			http.Error(w, "invoice_number must not be empty", http.StatusBadRequest)
			return
		}
		if invoiceUpdate.InvoiceDate.IsZero() {
			http.Error(w, "invoice_date must be provided", http.StatusBadRequest)
			return
		}
		if invoiceUpdate.CustomerID <= 0 {
			http.Error(w, "customer_id should be a positive number", http.StatusBadRequest)
			return
		}

		updatedInvoice, err := h.Queries.UpdateInvoice(r.Context(), database.UpdateInvoiceParams{
			ID:            int32(invoiceID),
			InvoiceNumber: invoiceUpdate.InvoiceNumber,
			InvoiceDate:   invoiceUpdate.InvoiceDate,
			CustomerID:    invoiceUpdate.CustomerID,
		})
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) {
				switch pqErr.Code {
				case "23505":
					// Unique constraint violation
					http.Error(w, "Invoice number must be unique", http.StatusConflict)
					return
				case "23503":
					// Foreign key violation
					http.Error(w, "Specified customer does not exist", http.StatusBadRequest)
					return
				default:
					writeInternalServerError(w, err)
					return
				}
			}
			writeInternalServerError(w, err)
			return
		}
		if updatedInvoice.Result != "success" {
			switch updatedInvoice.Result {
			case "invoice_not_found":
				http.Error(w, "Invoice not found", http.StatusNotFound)
				return
			default:
				writeInternalServerError(w, err)
				return
			}
		}
		writeServerResponse(w, http.StatusOK, invoiceResponse{
			ID:            updatedInvoice.ID.Int32,
			InvoiceNumber: updatedInvoice.InvoiceNumber.String,
			InvoiceDate:   updatedInvoice.InvoiceDate.Time,
			CustomerID:    updatedInvoice.CustomerID.Int32,
		})
	case http.MethodDelete:
		// DELETE /invoices/{invoice_id}
		deletionResult, err := h.Queries.DeleteInvoice(r.Context(), int32(invoiceID))
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) {
				// Check if it's a foreign key violation
				if pqErr.Code == "23503" { // 23503 is the SQLSTATE code for foreign key violation
					// Check the constraint name
					if pqErr.Constraint == "invoice_item_invoice_id_fkey" {
						http.Error(w, "cannot delete invoice: invoice is referenced in the invoice_item table", http.StatusConflict)
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
		if deletionResult == "invoice_not_found" {
			http.Error(w, "Invoice not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, config.MethodNotAllowedMsg, http.StatusMethodNotAllowed)
	}
}
