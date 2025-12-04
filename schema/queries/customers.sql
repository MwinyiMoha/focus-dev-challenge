-- name: GetCustomerById :one
SELECT * FROM customers WHERE id = @customer_id;
