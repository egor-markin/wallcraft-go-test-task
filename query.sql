------------------------------------------------------------------------------------------------------------------------
-- product
------------------------------------------------------------------------------------------------------------------------

-- name: ListProducts :many
SELECT * FROM product ORDER BY id LIMIT 100;

-- name: GetProduct :one
SELECT * FROM product WHERE id = $1;

-- name: CreateProduct :one
INSERT INTO product (name, description, price, available_items)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateProduct :one
UPDATE product
SET
    name = $2,
    description = $3,
    price = $4,
    available_items = $5
WHERE id = $1
RETURNING *;

-- name: DeleteProduct :one
WITH check_product AS (
    SELECT EXISTS(SELECT 1 FROM product WHERE id = @product_id::int) AS product_exists
),
delete_product AS (
    DELETE FROM product
    WHERE id = @product_id::int
    RETURNING *
)
SELECT
    CASE
        WHEN NOT (SELECT product_exists FROM check_product) THEN 'product_not_found'
        WHEN NOT EXISTS (SELECT 1 FROM delete_product) THEN 'delete_failed'
        ELSE 'success'
    END AS result
FROM delete_product
RIGHT JOIN (SELECT NULL) AS dummy ON true;

------------------------------------------------------------------------------------------------------------------------
-- invoice
------------------------------------------------------------------------------------------------------------------------

-- name: ListInvoices :many
SELECT * FROM invoice ORDER BY id LIMIT 100;

-- name: GetInvoice :one
SELECT * FROM invoice WHERE id = $1;

-- name: CreateInvoice :one
INSERT INTO invoice (invoice_number, invoice_date, customer_id)
VALUES (@invoice_number::text, @invoice_date::timestamp, @customer_id::int)
RETURNING *;

-- name: UpdateInvoice :one
WITH
    check_invoice AS (
        SELECT EXISTS(SELECT 1 FROM invoice i WHERE i.id = $1) AS invoice_exists
    ),
    update_invoice AS (
        UPDATE invoice
        SET
            invoice_number = @invoice_number::text,
            invoice_date = @invoice_date::timestamp,
            customer_id = @customer_id::int
        WHERE id = $1
        RETURNING *
    )
SELECT
    CASE
        WHEN NOT (SELECT invoice_exists FROM check_invoice) THEN 'invoice_not_found'
        WHEN NOT EXISTS (SELECT 1 FROM update_invoice) THEN 'update_failed'
        ELSE 'success'
    END AS result,
    update_invoice.*
FROM update_invoice
RIGHT JOIN (SELECT NULL) AS dummy ON true;

-- name: DeleteInvoice :one
WITH check_invoice AS (
    SELECT EXISTS(SELECT 1 FROM invoice WHERE id = @invoice_id::int) AS invoice_exists
),
delete_invoice AS (
    DELETE FROM invoice
    WHERE id = @invoice_id::int
    RETURNING *
)
SELECT
    CASE
        WHEN NOT (SELECT invoice_exists FROM check_invoice) THEN 'invoice_not_found'
        WHEN NOT EXISTS (SELECT 1 FROM delete_invoice) THEN 'delete_failed'
        ELSE 'success'
    END AS result
FROM delete_invoice
RIGHT JOIN (SELECT NULL) AS dummy ON true;

------------------------------------------------------------------------------------------------------------------------
-- customer
------------------------------------------------------------------------------------------------------------------------

-- name: ListCustomers :many
SELECT * FROM customer ORDER BY id LIMIT 100;

-- name: GetCustomer :one
SELECT * FROM customer WHERE id = $1;

-- name: CreateCustomer :one
INSERT INTO customer (first_name, last_name)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateCustomer :one
UPDATE customer
SET
    first_name = $2,
    last_name = $3
WHERE id = $1
RETURNING *;

-- name: DeleteCustomer :one
WITH check_customer AS (
    SELECT EXISTS(SELECT 1 FROM customer WHERE id = @customer_id::int) AS customer_exists
),
delete_customer AS (
    DELETE FROM customer
    WHERE id = @customer_id::int
    RETURNING *
)
SELECT
    CASE
        WHEN NOT (SELECT customer_exists FROM check_customer) THEN 'customer_not_found'
        WHEN NOT EXISTS (SELECT 1 FROM delete_customer) THEN 'delete_failed'
        ELSE 'success'
    END AS result
FROM delete_customer
RIGHT JOIN (SELECT NULL) AS dummy ON true;

------------------------------------------------------------------------------------------------------------------------
-- invoice_item
------------------------------------------------------------------------------------------------------------------------

-- name: ListProductsFromInvoice :many
SELECT
    p.id,
    p.name,
    p.description,
    p.price,
    ii.count,
    CAST((p.price * ii.count) AS numeric(10,2)) AS sum
FROM
    invoice_item ii
    JOIN Product p ON ii.product_id = p.id
WHERE
    ii.invoice_id = $1
ORDER BY
    p.id
 LIMIT
    100;

-- name: AddProductToInvoice :one
INSERT INTO invoice_item (invoice_id, product_id, count)
VALUES (@invoice_id::int, @product_id::int, @count::int)
ON CONFLICT (invoice_id, product_id)
DO UPDATE SET
    count = EXCLUDED.count
RETURNING *;

-- name: DeleteProductFromInvoice :one
WITH
    check_invoice_item AS (
        SELECT EXISTS(
            SELECT 1 FROM invoice_item
            WHERE invoice_id = @invoice_id::int AND product_id = @product_id::int
        ) AS invoice_item_exists
    ),
    delete_invoice_item AS (
        DELETE FROM invoice_item
        WHERE invoice_id = @invoice_id::int AND product_id = @product_id::int
        RETURNING *
    )
SELECT
    CASE
        WHEN NOT (SELECT invoice_item_exists FROM check_invoice_item) THEN 'invoice_item_not_found'
        WHEN NOT EXISTS (SELECT 1 FROM delete_invoice_item) THEN 'delete_failed'
        ELSE 'success'
    END AS result
FROM delete_invoice_item
RIGHT JOIN (SELECT NULL) AS dummy ON true;
