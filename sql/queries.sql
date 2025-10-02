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

-- name: DeleteSubscription :execrows
DELETE FROM subscriptions WHERE id = $1;

-- name: ListSubscriptions :many
SELECT * FROM subscriptions
WHERE 
    (sqlc.narg('user_id')::UUID IS NULL OR user_id = sqlc.narg('user_id')) AND
    (sqlc.narg('service_name')::VARCHAR IS NULL OR service_name ILIKE '%' || sqlc.narg('service_name') || '%')
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountSubscriptions :one
SELECT COUNT(*) FROM subscriptions
WHERE 
    (sqlc.narg('user_id')::UUID IS NULL OR user_id = sqlc.narg('user_id')) AND
    (sqlc.narg('service_name')::VARCHAR IS NULL OR service_name ILIKE '%' || sqlc.narg('service_name') || '%');

-- name: CalculateTotalCost :one
WITH date_range AS (
    SELECT 
        generate_series(sqlc.arg('start_date')::DATE, sqlc.arg('end_date')::DATE, '1 month'::interval)::DATE AS month_start
),
subscription_costs AS (
    SELECT 
        s.id as subscription_id,
        s.price,
        COUNT(DISTINCT dr.month_start) as months_count
    FROM subscriptions s
    CROSS JOIN date_range dr
    WHERE 
        (sqlc.narg('user_id')::UUID IS NULL OR s.user_id = sqlc.narg('user_id')) AND
        (sqlc.narg('service_name')::VARCHAR IS NULL OR s.service_name ILIKE '%' || sqlc.narg('service_name') || '%') AND
        (s.start_date <= dr.month_start) AND
        (s.end_date IS NULL OR s.end_date >= dr.month_start)
    GROUP BY s.id, s.price
)
SELECT COALESCE(SUM(price * months_count), 0)::BIGINT as total_cost
FROM subscription_costs;