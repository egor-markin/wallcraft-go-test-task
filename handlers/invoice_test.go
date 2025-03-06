package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/egor-markin/wallcraft-go-test-task/config"
	"github.com/egor-markin/wallcraft-go-test-task/database"
)

type invoiceMockQueries struct {
	ListInvoicesFunc             func(ctx context.Context) ([]database.Invoice, error)
	CreateInvoiceFunc            func(ctx context.Context, params database.CreateInvoiceParams) (database.Invoice, error)
	GetInvoiceFunc               func(ctx context.Context, id int32) (database.Invoice, error)
	UpdateInvoiceFunc            func(ctx context.Context, params database.UpdateInvoiceParams) (database.UpdateInvoiceRow, error)
	DeleteInvoiceFunc            func(ctx context.Context, id int32) (string, error)
	ListProductsFromInvoiceFunc  func(ctx context.Context, invoiceID int32) ([]database.ListProductsFromInvoiceRow, error)
	AddProductToInvoiceFunc      func(ctx context.Context, params database.AddProductToInvoiceParams) (database.InvoiceItem, error)
	DeleteProductFromInvoiceFunc func(ctx context.Context, params database.DeleteProductFromInvoiceParams) (string, error)
}

func (m *invoiceMockQueries) ListInvoices(ctx context.Context) ([]database.Invoice, error) {
	return m.ListInvoicesFunc(ctx)
}

func (m *invoiceMockQueries) CreateInvoice(ctx context.Context, params database.CreateInvoiceParams) (database.Invoice, error) {
	return m.CreateInvoiceFunc(ctx, params)
}

func (m *invoiceMockQueries) GetInvoice(ctx context.Context, id int32) (database.Invoice, error) {
	return m.GetInvoiceFunc(ctx, id)
}

func (m *invoiceMockQueries) UpdateInvoice(ctx context.Context, params database.UpdateInvoiceParams) (database.UpdateInvoiceRow, error) {
	return m.UpdateInvoiceFunc(ctx, params)
}

func (m *invoiceMockQueries) DeleteInvoice(ctx context.Context, id int32) (string, error) {
	return m.DeleteInvoiceFunc(ctx, id)
}

func (m *invoiceMockQueries) ListProductsFromInvoice(ctx context.Context, invoiceID int32) ([]database.ListProductsFromInvoiceRow, error) {
	return m.ListProductsFromInvoiceFunc(ctx, invoiceID)
}

func (m *invoiceMockQueries) AddProductToInvoice(ctx context.Context, params database.AddProductToInvoiceParams) (database.InvoiceItem, error) {
	return m.AddProductToInvoiceFunc(ctx, params)
}

func (m *invoiceMockQueries) DeleteProductFromInvoice(ctx context.Context, params database.DeleteProductFromInvoiceParams) (string, error) {
	return m.DeleteProductFromInvoiceFunc(ctx, params)
}

func TestInvoicesHandler(t *testing.T) {
	mockQueries := &invoiceMockQueries{}
	handler := &InvoiceHandler{Queries: mockQueries}

	t.Run("GET invoices - Success", func(t *testing.T) {
		mockQueries.ListInvoicesFunc = func(ctx context.Context) ([]database.Invoice, error) {
			now := time.Now().UTC()
			return []database.Invoice{
				{ID: 1, InvoiceNumber: "INV-001", InvoiceDate: now, CustomerID: 10},
				{ID: 2, InvoiceNumber: "INV-002", InvoiceDate: now, CustomerID: 20},
			}, nil
		}

		req := httptest.NewRequest(http.MethodGet, config.InvoicesApiPrefix, nil)
		w := httptest.NewRecorder()

		handler.InvoicesHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var invoices []invoiceResponse
		if err := json.Unmarshal(w.Body.Bytes(), &invoices); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if len(invoices) != 2 {
			t.Errorf("expected 2 invoices, got %d", len(invoices))
		}

		if invoices[0].InvoiceNumber != "INV-001" || invoices[1].InvoiceNumber != "INV-002" {
			t.Errorf("unexpected invoice numbers: %v", invoices)
		}
	})

	t.Run("POST invoices - Success", func(t *testing.T) {
		newInvoice := createInvoiceRequest{
			InvoiceNumber: "INV-003",
			CustomerID:    30,
		}

		mockQueries.CreateInvoiceFunc = func(ctx context.Context, params database.CreateInvoiceParams) (database.Invoice, error) {
			return database.Invoice{
				ID:            3,
				InvoiceNumber: newInvoice.InvoiceNumber,
				InvoiceDate:   time.Now().UTC(),
				CustomerID:    newInvoice.CustomerID,
			}, nil
		}

		invoiceJSON, _ := json.Marshal(newInvoice)
		req := httptest.NewRequest(http.MethodPost, config.InvoicesApiPrefix, bytes.NewBuffer(invoiceJSON))
		w := httptest.NewRecorder()

		handler.InvoicesHandler(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status code %d, got %d", http.StatusCreated, w.Code)
		}

		var createdInvoice invoiceResponse
		if err := json.Unmarshal(w.Body.Bytes(), &createdInvoice); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if createdInvoice.ID <= 0 || createdInvoice.InvoiceNumber != newInvoice.InvoiceNumber || createdInvoice.CustomerID != newInvoice.CustomerID {
			t.Errorf("unexpected created invoice: %v", createdInvoice)
		}
	})
}

func TestInvoiceHandler(t *testing.T) {
	mockQueries := &invoiceMockQueries{}
	handler := &InvoiceHandler{Queries: mockQueries}

	t.Run("GET invoices/{id} - Success", func(t *testing.T) {
		inv := database.Invoice{
			ID:            33,
			InvoiceNumber: "INV-033",
			InvoiceDate:   time.Now().UTC(),
			CustomerID:    100,
		}

		mockQueries.GetInvoiceFunc = func(ctx context.Context, id int32) (database.Invoice, error) {
			if id != inv.ID {
				return database.Invoice{}, sql.ErrNoRows
			}
			return inv, nil
		}

		req := httptest.NewRequest(http.MethodGet, config.InvoicesApiPrefix+"/"+strconv.Itoa(int(inv.ID)), nil)
		w := httptest.NewRecorder()

		handler.InvoiceHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var invoice invoiceResponse
		if err := json.Unmarshal(w.Body.Bytes(), &invoice); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if invoice.ID != inv.ID || invoice.InvoiceNumber != inv.InvoiceNumber || invoice.CustomerID != inv.CustomerID {
			t.Errorf("unexpected invoice: %v", invoice)
		}
	})

	t.Run("GET invoices/{id} - Not Found", func(t *testing.T) {
		mockQueries.GetInvoiceFunc = func(ctx context.Context, id int32) (database.Invoice, error) {
			return database.Invoice{}, sql.ErrNoRows
		}

		req := httptest.NewRequest(http.MethodGet, config.InvoicesApiPrefix+"/1", nil)
		w := httptest.NewRecorder()

		handler.InvoiceHandler(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status code %d, got %d", http.StatusNotFound, w.Code)
		}

		if w.Body.String() != "Invoice not found\n" {
			t.Errorf("unexpected response body: %s", w.Body.String())
		}
	})

	t.Run("PATCH invoices/{id} - Success", func(t *testing.T) {
		invoiceID := int32(24)
		updateParams := updateInvoiceRequest{
			InvoiceNumber: "INV-UPDATED",
			InvoiceDate:   time.Date(2025, time.March, 6, 15, 4, 5, 0, time.UTC),
			CustomerID:    50,
		}
		mockQueries.UpdateInvoiceFunc = func(ctx context.Context, params database.UpdateInvoiceParams) (database.UpdateInvoiceRow, error) {
			if params.ID != invoiceID {
				return database.UpdateInvoiceRow{}, errors.New("unexpected invoice ID")
			}
			return database.UpdateInvoiceRow{
				Result:        "success",
				ID:            sql.NullInt32{Int32: invoiceID, Valid: true},
				InvoiceNumber: sql.NullString{String: updateParams.InvoiceNumber, Valid: true},
				InvoiceDate:   sql.NullTime{Time: updateParams.InvoiceDate, Valid: true},
				CustomerID:    sql.NullInt32{Int32: updateParams.CustomerID, Valid: true},
			}, nil
		}

		updateJSON, _ := json.Marshal(updateParams)
		req := httptest.NewRequest(http.MethodPatch, config.InvoicesApiPrefix+"/"+strconv.Itoa(int(invoiceID)), bytes.NewBuffer(updateJSON))
		w := httptest.NewRecorder()

		handler.InvoiceHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var updatedInvoice invoiceResponse
		if err := json.Unmarshal(w.Body.Bytes(), &updatedInvoice); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if updatedInvoice.ID != invoiceID || updatedInvoice.InvoiceNumber != updateParams.InvoiceNumber || updatedInvoice.InvoiceDate != updateParams.InvoiceDate || updatedInvoice.CustomerID != updateParams.CustomerID {
			t.Errorf("unexpected updated invoice: %v", updatedInvoice)
		}
	})

	t.Run("DELETE invoices/{id} - Success", func(t *testing.T) {
		var invoiceID int32 = 444
		mockQueries.DeleteInvoiceFunc = func(ctx context.Context, id int32) (string, error) {
			if id != invoiceID {
				return "invoice_not_found", nil
			}
			return "success", nil
		}

		req := httptest.NewRequest(http.MethodDelete, config.InvoicesApiPrefix+"/"+strconv.Itoa(int(invoiceID)), nil)
		w := httptest.NewRecorder()

		handler.InvoiceHandler(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status code %d, got %d", http.StatusNoContent, w.Code)
		}
	})
}

func TestInvoiceItemHandler(t *testing.T) {
	mockQueries := &invoiceMockQueries{}
	handler := &InvoiceHandler{Queries: mockQueries}

	// GET /invoices/{invoice_id}/products
	t.Run("GET invoice items - Success", func(t *testing.T) {
		mockInvoiceID := int32(45)
		list := []database.ListProductsFromInvoiceRow{
			{ID: 1, Name: "Product 1", Price: "100.0", Count: 2},
			{ID: 2, Name: "Product 2", Price: "300.0", Count: 4},
		}
		mockQueries.ListProductsFromInvoiceFunc = func(ctx context.Context, invoiceID int32) ([]database.ListProductsFromInvoiceRow, error) {
			if invoiceID != mockInvoiceID {
				return nil, sql.ErrNoRows
			}
			return list, nil
		}

		req := httptest.NewRequest(http.MethodGet, config.InvoicesApiPrefix+"/"+strconv.Itoa(int(mockInvoiceID))+"/products", nil)
		w := httptest.NewRecorder()

		handler.InvoiceHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var fetchedProducts []invoiceProductResponse
		if err := json.Unmarshal(w.Body.Bytes(), &fetchedProducts); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if len(fetchedProducts) != len(list) {
			t.Errorf("expected 2 products, got %d", len(list))
		}
		if fetchedProducts[0].Name != list[0].Name || fetchedProducts[1].Name != list[1].Name {
			t.Errorf("unexpected product names: %v", fetchedProducts)
		}

	})

	// POST /invoices/{invoice_id}/products
	t.Run("POST invoice items - Success", func(t *testing.T) {
		mockInvoiceID := int32(98)
		mockProductID := int32(99)
		mockCount := int32(24)
		params := createInvoiceItemRequest{Count: mockCount}
		mockQueries.AddProductToInvoiceFunc = func(ctx context.Context, p database.AddProductToInvoiceParams) (database.InvoiceItem, error) {
			if p.InvoiceID != mockInvoiceID {
				return database.InvoiceItem{}, errors.New("unexpected invoice ID")
			}
			if p.ProductID != mockProductID {
				return database.InvoiceItem{}, errors.New("unexpected product ID")
			}
			if p.Count != mockCount {
				return database.InvoiceItem{}, errors.New("unexpected count")
			}
			return database.InvoiceItem{ID: 1, InvoiceID: p.InvoiceID, ProductID: p.ProductID, Count: p.Count}, nil
		}

		paramsJSON, _ := json.Marshal(params)
		req := httptest.NewRequest(http.MethodPost, config.InvoicesApiPrefix+"/"+strconv.Itoa(int(mockInvoiceID))+"/products/"+strconv.Itoa(int(mockProductID)), bytes.NewBuffer(paramsJSON))
		w := httptest.NewRecorder()

		handler.InvoiceHandler(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status code %d, got %d", http.StatusCreated, w.Code)
		}

		var createdInvoiceItem invoiceItemResponse
		if err := json.Unmarshal(w.Body.Bytes(), &createdInvoiceItem); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if createdInvoiceItem.ID <= 0 || createdInvoiceItem.InvoiceID != mockInvoiceID || createdInvoiceItem.ProductID != mockProductID || createdInvoiceItem.Count != params.Count {
			t.Errorf("unexpected created product: %v", createdInvoiceItem)
		}

	})

	// DELETE /invoices/{invoice_id}/products/{product_id}
	t.Run("DELETE invoice items - Success", func(t *testing.T) {
		var mockInvoiceID int32 = 678
		var mockProductID int32 = 345
		mockQueries.DeleteProductFromInvoiceFunc = func(ctx context.Context, params database.DeleteProductFromInvoiceParams) (string, error) {
			if params.InvoiceID != mockInvoiceID {
				return "", errors.New("unexpected invoice ID")
			}
			if params.ProductID != mockProductID {
				return "", errors.New("unexpected product ID")
			}
			return "success", nil
		}

		req := httptest.NewRequest(http.MethodDelete, config.InvoicesApiPrefix+"/"+strconv.Itoa(int(mockInvoiceID))+"/products/"+strconv.Itoa(int(mockProductID)), nil)
		w := httptest.NewRecorder()

		handler.InvoiceHandler(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status code %d, got %d", http.StatusNoContent, w.Code)
		}
	})
}
