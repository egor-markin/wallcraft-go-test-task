# Wallcraft Golang Test Task

This project is a RESTful API built in Go, designed to manage products, customers, invoices, and invoice items.

Below you will find instructions on how to build and run the application, as well as details on the available endpoints and how to use them.

## Build and run instructions

### Building the App Binary
To build the application binary, run the following command in the root directory of the project:
```bash
go build -o wallcraft-go-test-task
```
This will generate an executable named `wallcraft-go-test-task`.

You can run it using the following command, replacing Postgresql credentials:
```bash
DATABASE_URL="postgres://user:password@db:5432/mydb?sslmode=disable" ./wallcraft-go-test-task
```

### Building the Docker Image
To build the Docker image, ensure you have Docker installed and run the following command in the root directory of the project:
```bash
docker build -t wallcraft-go-test-task:latest .
```
This will create a Docker image tagged as `wallcraft-go-test-task:latest`.

After you can run it as a single docker container:
```bash
docker run -d \
  --name wallcraft-app \
  -e DATABASE_URL="postgres://user:password@db:5432/mydb?sslmode=disable" \
  -p 8080:8080 \
  wallcraft-go-test-task:latest
```

To run the application using Docker Compose, ensure you have Docker Compose installed and run the following command in the root directory of the project:
```bash
docker compose up -d
```
This will start the application and the PostgreSQL database. The API will be available at http://localhost:8080.

## Configuration

The application requires the following environment variable:
- DATABASE_URL: The connection string for the PostgreSQL database. Example: `postgres://user:password@db:5432/mydb?sslmode=disable`. At this moment only Postgresql database is supported.

## Database Schema

The database schema is defined in the `schema.sql` file. It includes tables for customer, product, invoice, and invoice_item.

## API Endpoints

### Products

#### GET /api/v1/products
Returns a list of products (limited to the first 100 items).

Example Request:
```bash
curl --location 'http://localhost:8080/api/v1/products'
```
Example Response:
```json
[
    {
        "id": 1,
        "name": "Mouse",
        "description": "Optical Logitech mouse with 1000dpi",
        "price": "222.00",
        "available_items": 22
    }
]
```

#### GET /api/v1/products/{product_id}
Returns a single product or status 404 if none is found.

Example Request:

````bash
curl --location 'http://localhost:8080/api/v1/products/1'
````

#### POST /api/v1/products
Creates a new product.

Example Request:
```bash
curl --location 'http://localhost:8080/api/v1/products' \
--header 'Content-Type: application/json' \
--data '{
    "name": "Mouse",
    "description": "Optical Logitech mouse with 1000dpi",
    "price": "222",
    "available_items": 22
}'
```
Example Response:
```json
{
    "id": 1,
    "name": "Mouse",
    "description": "Optical Logitech mouse with 1000dpi",
    "price": "222.00",
    "available_items": 22
}
```
#### PATCH /api/v1/products/{product_id}
Updates an existing product.

Example Request:
```bash
curl --location --request PATCH 'http://localhost:8080/api/v1/products/2' \
--header 'Content-Type: application/json' \
--data '{
    "name": "Keyboard",
    "description": "Mechanical Cherry keyboard",
    "price": "50.21",
    "available_items": 33
}'
```

Example Response:
```json
{
    "id": 2,
    "name": "Keyboard",
    "description": "Mechanical Cherry keyboard",
    "price": "50.21",
    "available_items": 33
}
```

#### DELETE /api/v1/products/{product_id}
Deletes a product. Returns 204 with an empty body for success or 404 if the product wasn't found. If there are related invoice items, 409 Conflict Status is returned.

Example Request:
```bash
curl --location --request DELETE 'http://localhost:8080/api/v1/products/7'
```
### Customers

#### GET /api/v1/customers
Returns a list of customers (limited to the first 100 items).

Example Request:
```bash
curl --location 'http://localhost:8080/api/v1/customers'
```
Example Response:
```json
[
    {
        "id": 1,
        "first_name": "Jarred",
        "last_name": "Black"
    }
]
```
#### GET /api/v1/customers/{customer_id}
Returns a single customer or status 404 if none is found.

Example Request:
```bash
curl --location 'http://localhost:8080/api/v1/customers/1'
```

#### POST /api/v1/customers
Creates a new customer.

Example Request:
```bash
curl --location 'http://localhost:8080/api/v1/customers' \
--header 'Content-Type: application/json' \
--data '{
    "first_name": "Jarred",
    "last_name": "Black"
}'
```
Example Response:
```json
{
    "id": 2,
    "first_name": "Jarred",
    "last_name": "Black"
}
```

#### PATCH /api/v1/customers/{customer_id}
Updates an existing customer.

Example Request:
```bash
curl --location --request PATCH 'http://localhost:8080/api/v1/customers/9' \
--header 'Content-Type: application/json' \
--data '{
    "first_name": "Joe",
    "last_name": "White"
}'
```
Example Response:
```json
{
    "id": 1,
    "first_name": "Joe",
    "last_name": "White"
}
```

#### DELETE /api/v1/customers/{customer_id}
Deletes a customer. Returns 204 with an empty body for success or 404 if the customer wasn't found. If there are related invoices, 409 Conflict Status is returned.

Example Request:
```bash
curl --location --request DELETE 'http://localhost:8080/api/v1/customers/1'
```

### Invoices

#### GET /api/v1/invoices
Returns a list of invoices (limited to the first 100 items).

Example Request:
```bash
curl --location 'http://localhost:8080/api/v1/invoices'
```

Example Response:
```json
[
    {
        "id": 1,
        "invoice_number": "INV-33318",
        "invoice_date": "2025-03-06T10:20:58.521504Z",
        "customer_id": 1
    }
]
```

#### GET /api/v1/invoices/{invoice_id}
Returns a single invoice or status 404 if none is found.
Example Request:
```bash
curl --location 'http://localhost:8080/api/v1/invoices/2'
```

#### POST /api/v1/invoices
Creates a new invoice.

Example Request:
```bash
curl --location 'http://localhost:8080/api/v1/invoices' \
--header 'Content-Type: application/json' \
--data '{
    "invoice_number": "INV-33318",
    "invoice_date": "2025-03-06T10:20:58.521504Z",
    "customer_id": 1
}'
```

Example Response:
```json
{
    "id": 1,
    "invoice_number": "INV-33318",
    "invoice_date": "2025-03-06T10:20:58.521504Z",
    "customer_id": 1
}
```
#### PATCH /api/v1/invoices/{invoice_id}
Updates an existing invoice.

Example Request:
```bash
curl --location --request PATCH 'http://localhost:8080/api/v1/invoices/1' \
--header 'Content-Type: application/json' \
--data '{
    "invoice_number": "INV-322342",
    "invoice_date": "2025-06-22T14:33:12Z",
    "customer_id": 1
}'
```

Example Response:
```json
{
    "id": 1,
    "invoice_number": "INV-322342",
    "invoice_date": "2025-06-22T14:33:12Z",
    "customer_id": 1
}
```

#### DELETE /api/v1/invoices/{invoice_id}
Deletes an invoice. Returns 204 with an empty body for success or 404 if the invoice wasn't found. If there are related invoice items, 409 Conflict Status is returned.

Example Request:
```bash
curl --location --request DELETE 'http://localhost:8080/api/v1/invoices/1'
```

### Invoice Products

#### GET /api/v1/invoices/{invoice_id}/products
Returns a list of products that belong to the provided invoice (limited to the first 100 items).

Example Request:
```bash
curl --location 'http://localhost:8080/api/v1/invoices/1/products'
```
Example Response:
```json
[
    {
        "id": 2,
        "name": "Keyboard",
        "description": "Mechanical Cherry keyboard",
        "price": "50.21",
        "count": 5,
        "sum": "251.05"
    }
]
```

#### POST /api/v1/invoices/{invoice_id}/products/{product_id}
Adds a product to an invoice.

Example Request:
```bash
curl --location 'http://localhost:8080/api/v1/invoices/2/products/2' \
--header 'Content-Type: application/json' \
--data '{
    "count": 5
}'
```
Example Response:
```json
{
    "id": 4,
    "invoice_id": 2,
    "product_id": 2,
    "count": 5
}
```

#### DELETE /api/v1/invoices/{invoice_id}/products/{product_id}
Deletes a product from an invoice. Returns 204 No Content status for success.

Example Request:
```bash
curl --location --request DELETE 'http://localhost:8080/api/v1/invoices/1/products/1'
```

### Health Check GET /api/v1/health
Health check endpoint for Docker Compose, Kubernetes, etc. Returns "OK" with status 200.

Example Request:
```bash
curl --location 'http://localhost:8080/api/v1/health'
```
Example Response:
```
OK
```

## SQLC Code Generation

This project uses [SQLC](https://sqlc.dev/) to generate type-safe Go code from SQL queries. Below are the steps to generate the Go code.

### Prerequisites
- Install sqlc by following the official installation guide: [SQLC Installation](https://docs.sqlc.dev/en/latest/overview/install.html).
- Ensure you have a sqlc.yaml configuration file in your project root.

### Generating Go Code
- Make sure that `schema.sql` and `query.sql` files are in the project root.
- Run the following command to generate the Go code:
```bash
sqlc generate
```
This will generate Go code in the `database` directory based on the SQL schema and queries.
