package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/egor-markin/wallcraft-go-test-task/config"
	"github.com/egor-markin/wallcraft-go-test-task/database"
)

// customerMockQueries implements the CustomerQueries interface for testing.
type customerMockQueries struct {
	ListCustomersFunc  func(ctx context.Context) ([]database.Customer, error)
	CreateCustomerFunc func(ctx context.Context, params database.CreateCustomerParams) (database.Customer, error)
	GetCustomerFunc    func(ctx context.Context, id int32) (database.Customer, error)
	UpdateCustomerFunc func(ctx context.Context, params database.UpdateCustomerParams) (database.Customer, error)
	DeleteCustomerFunc func(ctx context.Context, id int32) (string, error)
}

func (m *customerMockQueries) ListCustomers(ctx context.Context) ([]database.Customer, error) {
	return m.ListCustomersFunc(ctx)
}

func (m *customerMockQueries) CreateCustomer(ctx context.Context, params database.CreateCustomerParams) (database.Customer, error) {
	return m.CreateCustomerFunc(ctx, params)
}

func (m *customerMockQueries) GetCustomer(ctx context.Context, id int32) (database.Customer, error) {
	return m.GetCustomerFunc(ctx, id)
}

func (m *customerMockQueries) UpdateCustomer(ctx context.Context, params database.UpdateCustomerParams) (database.Customer, error) {
	return m.UpdateCustomerFunc(ctx, params)
}

func (m *customerMockQueries) DeleteCustomer(ctx context.Context, id int32) (string, error) {
	return m.DeleteCustomerFunc(ctx, id)
}

func TestCustomersHandler(t *testing.T) {
	mockQueries := &customerMockQueries{}
	handler := &CustomerHandler{Queries: mockQueries}

	t.Run("GET customers - Success", func(t *testing.T) {
		mockQueries.ListCustomersFunc = func(ctx context.Context) ([]database.Customer, error) {
			return []database.Customer{
				{ID: 1, FirstName: "John", LastName: "Doe"},
				{ID: 2, FirstName: "Jane", LastName: "Smith"},
			}, nil
		}

		req := httptest.NewRequest(http.MethodGet, config.CustomersApiPrefix, nil)
		w := httptest.NewRecorder()

		handler.CustomersHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var customers []customerResponse
		if err := json.Unmarshal(w.Body.Bytes(), &customers); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if len(customers) != 2 {
			t.Errorf("expected 2 customers, got %d", len(customers))
		}

		if customers[0].FirstName != "John" || customers[1].FirstName != "Jane" {
			t.Errorf("unexpected customer names: %v", customers)
		}
	})

	t.Run("POST customers - Success", func(t *testing.T) {
		newCustomer := createCustomerRequest{FirstName: "Alice", LastName: "Wonderland"}

		mockQueries.CreateCustomerFunc = func(ctx context.Context, params database.CreateCustomerParams) (database.Customer, error) {
			return database.Customer{ID: 3, FirstName: newCustomer.FirstName, LastName: newCustomer.LastName}, nil
		}

		customerJSON, _ := json.Marshal(newCustomer)
		req := httptest.NewRequest(http.MethodPost, config.CustomersApiPrefix, bytes.NewBuffer(customerJSON))
		w := httptest.NewRecorder()

		handler.CustomersHandler(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status code %d, got %d", http.StatusCreated, w.Code)
		}

		var createdCustomer customerResponse
		if err := json.Unmarshal(w.Body.Bytes(), &createdCustomer); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if createdCustomer.ID <= 0 || createdCustomer.FirstName != newCustomer.FirstName || createdCustomer.LastName != newCustomer.LastName {
			t.Errorf("unexpected created customer: %v", createdCustomer)
		}
	})
}

func TestCustomerHandler(t *testing.T) {
	mockQueries := &customerMockQueries{}
	handler := &CustomerHandler{Queries: mockQueries}

	t.Run("GET customers/{id} - Success", func(t *testing.T) {
		c := database.Customer{ID: 33, FirstName: "John", LastName: "Doe"}

		mockQueries.GetCustomerFunc = func(ctx context.Context, id int32) (database.Customer, error) {
			if id != c.ID {
				return database.Customer{}, sql.ErrNoRows
			}
			return database.Customer{ID: c.ID, FirstName: c.FirstName, LastName: c.LastName}, nil
		}

		req := httptest.NewRequest(http.MethodGet, config.CustomersApiPrefix+"/"+strconv.Itoa(int(c.ID)), nil)
		w := httptest.NewRecorder()

		handler.CustomerHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var customer customerResponse
		if err := json.Unmarshal(w.Body.Bytes(), &customer); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if customer.ID != c.ID || customer.FirstName != c.FirstName || customer.LastName != c.LastName {
			t.Errorf("unexpected customer: %v", customer)
		}
	})

	t.Run("GET customers/{id} - Not Found", func(t *testing.T) {
		mockQueries.GetCustomerFunc = func(ctx context.Context, id int32) (database.Customer, error) {
			return database.Customer{}, sql.ErrNoRows
		}

		req := httptest.NewRequest(http.MethodGet, config.CustomersApiPrefix+"/1", nil)
		w := httptest.NewRecorder()

		handler.CustomerHandler(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status code %d, got %d", http.StatusNotFound, w.Code)
		}

		if w.Body.String() != "Customer not found\n" {
			t.Errorf("unexpected response body: %s", w.Body.String())
		}
	})

	t.Run("PATCH customers/{id} - Success", func(t *testing.T) {
		customerId := int32(97)
		updateParams := updateCustomerRequest{FirstName: "Alice", LastName: "Cooper"}
		mockQueries.UpdateCustomerFunc = func(ctx context.Context, params database.UpdateCustomerParams) (database.Customer, error) {
			if params.ID != customerId {
				return database.Customer{}, sql.ErrNoRows
			}
			return database.Customer{ID: customerId, FirstName: updateParams.FirstName, LastName: updateParams.LastName}, nil
		}

		updateJSON, _ := json.Marshal(updateParams)
		req := httptest.NewRequest(http.MethodPatch, config.CustomersApiPrefix+"/"+strconv.Itoa(int(customerId)), bytes.NewBuffer(updateJSON))
		w := httptest.NewRecorder()

		handler.CustomerHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var updatedCustomer customerResponse
		if err := json.Unmarshal(w.Body.Bytes(), &updatedCustomer); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if updatedCustomer.ID != customerId || updatedCustomer.FirstName != updateParams.FirstName || updatedCustomer.LastName != updateParams.LastName {
			t.Errorf("unexpected updated customer: %v", updatedCustomer)
		}
	})

	t.Run("DELETE customers/{id} - Success", func(t *testing.T) {
		var customerID int32 = 444
		mockQueries.DeleteCustomerFunc = func(ctx context.Context, id int32) (string, error) {
			if id != customerID {
				return "customer_not_found", nil
			}
			return "success", nil
		}

		req := httptest.NewRequest(http.MethodDelete, config.CustomersApiPrefix+"/"+strconv.Itoa(int(customerID)), nil)
		w := httptest.NewRecorder()

		handler.CustomerHandler(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status code %d, got %d", http.StatusNoContent, w.Code)
		}
	})
}
