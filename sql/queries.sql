-- name: CreateSubscription :one
INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetSubscription :one
SELECT * FROM subscriptions WHERE id = $1;

-- name: UpdateSubscription :one
UPDATE subscriptions 
SET 
    service_name = COALESCE($2, service_name),
    price = COALESCE($3, price),
    start_date = COALESCE($4, start_date),
    end_date = COALESCE($5, end_date),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteSubscription :exec
DELETE FROM subscriptions WHERE id = $1;

-- name: ListSubscriptions :many
SELECT * FROM subscriptions
WHERE 
    ($1::UUID IS NULL OR user_id = $1) AND
    ($2::VARCHAR IS NULL OR service_name ILIKE '%' || $2 || '%')
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountSubscriptions :one
SELECT COUNT(*) FROM subscriptions
WHERE 
    ($1::UUID IS NULL OR user_id = $1) AND
    ($2::VARCHAR IS NULL OR service_name ILIKE '%' || $2 || '%');

-- name: CalculateTotalCost :one
SELECT COALESCE(SUM(price), 0)::BIGINT as total_cost FROM subscriptions
WHERE 
    ($1::UUID IS NULL OR user_id = $1) AND
    ($2::VARCHAR IS NULL OR service_name ILIKE '%' || $2 || '%') AND
    (start_date <= $4) AND
    (end_date IS NULL OR end_date >= $3);