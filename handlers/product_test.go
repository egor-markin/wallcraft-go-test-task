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

type productMockQueries struct {
	ListProductsFunc  func(ctx context.Context) ([]database.Product, error)
	CreateProductFunc func(ctx context.Context, params database.CreateProductParams) (database.Product, error)
	GetProductFunc    func(ctx context.Context, id int32) (database.Product, error)
	UpdateProductFunc func(ctx context.Context, params database.UpdateProductParams) (database.Product, error)
	DeleteProductFunc func(ctx context.Context, id int32) (string, error)
	WithTxFunc        func(tx *sql.Tx) *database.Queries
}

func (m *productMockQueries) ListProducts(ctx context.Context) ([]database.Product, error) {
	return m.ListProductsFunc(ctx)
}

func (m *productMockQueries) CreateProduct(ctx context.Context, params database.CreateProductParams) (database.Product, error) {
	return m.CreateProductFunc(ctx, params)
}

func (m *productMockQueries) GetProduct(ctx context.Context, id int32) (database.Product, error) {
	return m.GetProductFunc(ctx, id)
}

func (m *productMockQueries) UpdateProduct(ctx context.Context, params database.UpdateProductParams) (database.Product, error) {
	return m.UpdateProductFunc(ctx, params)
}

func (m *productMockQueries) DeleteProduct(ctx context.Context, id int32) (string, error) {
	return m.DeleteProductFunc(ctx, id)
}

func (m *productMockQueries) WithTx(tx *sql.Tx) *database.Queries {
	return m.WithTxFunc(tx)
}

func TestProductsHandler(t *testing.T) {
	mockQueries := &productMockQueries{}
	handler := &ProductHandler{Queries: mockQueries}

	// GET /products
	t.Run("GET products - Success", func(t *testing.T) {
		mockQueries.ListProductsFunc = func(ctx context.Context) ([]database.Product, error) {
			return []database.Product{
				{ID: 1, Name: "Product 1", Price: "100.0"},
				{ID: 2, Name: "Product 2", Price: "200.0"},
			}, nil
		}

		req := httptest.NewRequest(http.MethodGet, config.ProductsApiPrefix, nil)
		w := httptest.NewRecorder()

		handler.ProductsHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var products []productResponse
		if err := json.Unmarshal(w.Body.Bytes(), &products); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if len(products) != 2 {
			t.Errorf("expected 2 products, got %d", len(products))
		}

		if products[0].Name != "Product 1" || products[1].Name != "Product 2" {
			t.Errorf("unexpected product names: %v", products)
		}
	})

	// POST /products
	t.Run("POST products - Success", func(t *testing.T) {
		newProduct := createProductRequest{Name: "New Product", Price: "150.0"}

		mockQueries.CreateProductFunc = func(ctx context.Context, params database.CreateProductParams) (database.Product, error) {
			return database.Product{ID: 3, Name: newProduct.Name, Price: newProduct.Price}, nil
		}

		productJSON, _ := json.Marshal(newProduct)
		req := httptest.NewRequest(http.MethodPost, config.ProductsApiPrefix, bytes.NewBuffer(productJSON))
		w := httptest.NewRecorder()

		handler.ProductsHandler(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status code %d, got %d", http.StatusCreated, w.Code)
		}

		var createdProduct productResponse
		if err := json.Unmarshal(w.Body.Bytes(), &createdProduct); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if createdProduct.ID <= 0 || createdProduct.Name != newProduct.Name || createdProduct.Price != newProduct.Price {
			t.Errorf("unexpected created product: %v", createdProduct)
		}
	})
}

func TestProductHandler(t *testing.T) {
	mockQueries := &productMockQueries{}
	handler := &ProductHandler{Queries: mockQueries}

	// GET /products/{id}
	t.Run("GET products/{id} - Success", func(t *testing.T) {
		p := database.Product{ID: 33, Name: "Product 1", Price: "333.3"}

		mockQueries.GetProductFunc = func(ctx context.Context, id int32) (database.Product, error) {
			if id != p.ID {
				return database.Product{}, sql.ErrNoRows
			}
			return p, nil
		}

		req := httptest.NewRequest(http.MethodGet, config.ProductsApiPrefix+"/"+strconv.Itoa(int(p.ID)), nil)
		w := httptest.NewRecorder()

		handler.ProductHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var product productResponse
		if err := json.Unmarshal(w.Body.Bytes(), &product); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if product.ID != p.ID || product.Name != p.Name || product.Price != p.Price {
			t.Errorf("unexpected product: %v", product)
		}
	})

	// GET /products/{id}
	t.Run("GET products/{id} - Not Found", func(t *testing.T) {
		mockQueries.GetProductFunc = func(ctx context.Context, id int32) (database.Product, error) {
			return database.Product{}, sql.ErrNoRows
		}

		req := httptest.NewRequest(http.MethodGet, config.ProductsApiPrefix+"/1", nil)
		w := httptest.NewRecorder()

		handler.ProductHandler(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status code %d, got %d", http.StatusNotFound, w.Code)
		}

		if w.Body.String() != "Product not found\n" {
			t.Errorf("unexpected response body: %s", w.Body.String())
		}
	})

	// PATCH /products/{id}
	t.Run("PATCH products/{id} - Success", func(t *testing.T) {
		productID := int32(123)
		updateParams := updateProductRequest{Name: "Updated Product", Price: "150.0"}
		mockQueries.UpdateProductFunc = func(ctx context.Context, params database.UpdateProductParams) (database.Product, error) {
			if params.ID != productID {
				return database.Product{}, sql.ErrNoRows
			}
			return database.Product{ID: productID, Name: updateParams.Name, Price: updateParams.Price}, nil
		}

		updateJSON, _ := json.Marshal(updateParams)
		req := httptest.NewRequest(http.MethodPatch, config.ProductsApiPrefix+"/"+strconv.Itoa(int(productID)), bytes.NewBuffer(updateJSON))
		w := httptest.NewRecorder()

		handler.ProductHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var updatedProduct productResponse
		if err := json.Unmarshal(w.Body.Bytes(), &updatedProduct); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if updatedProduct.ID != productID || updatedProduct.Name != updateParams.Name || updatedProduct.Price != updateParams.Price {
			t.Errorf("unexpected updated product: %v", updatedProduct)
		}
	})

	// DELETE products/{id}
	t.Run("DELETE products/{id} - Success", func(t *testing.T) {
		var productId int32 = 444
		mockQueries.DeleteProductFunc = func(ctx context.Context, id int32) (string, error) {
			if id != productId {
				return "product_not_found", nil
			}
			return "success", nil
		}

		req := httptest.NewRequest(http.MethodDelete, config.ProductsApiPrefix+"/"+strconv.Itoa(int(productId)), nil)
		w := httptest.NewRecorder()

		handler.ProductHandler(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status code %d, got %d", http.StatusNoContent, w.Code)
		}
	})
}
