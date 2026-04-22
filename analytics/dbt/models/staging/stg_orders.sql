-- Staging model: orders from commerce.orders table
-- Normalises raw order data for downstream use

WITH source AS (
    SELECT * FROM {{ source('commerce', 'orders') }}
),

renamed AS (
    SELECT
        id                                      AS order_id,
        user_id,
        status,
        LOWER(status)                           AS status_key,
        total_amount                            AS total_usd,
        subtotal_amount                         AS subtotal_usd,
        shipping_amount                         AS shipping_usd,
        tax_amount                              AS tax_usd,
        currency,
        coupon_code,
        shipping_country,
        payment_method,
        created_at                              AS ordered_at,
        updated_at,
        DATE_TRUNC('day',  created_at)::DATE    AS ordered_date,
        DATE_TRUNC('week', created_at)::DATE    AS ordered_week,
        DATE_TRUNC('month',created_at)::DATE    AS ordered_month,
        EXTRACT(HOUR FROM created_at)           AS ordered_hour
    FROM source
    WHERE id IS NOT NULL
      AND created_at >= '{{ var("start_date") }}'
)

SELECT * FROM renamed
