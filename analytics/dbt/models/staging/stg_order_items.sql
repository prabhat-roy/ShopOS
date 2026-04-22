WITH source AS (
    SELECT * FROM {{ source('commerce', 'order_items') }}
),

renamed AS (
    SELECT
        id              AS order_item_id,
        order_id,
        product_id,
        sku,
        quantity,
        unit_price_usd,
        total_price_usd,
        discount_amount_usd,
        category_id,
        brand_id,
        created_at
    FROM source
    WHERE order_id IS NOT NULL
)

SELECT * FROM renamed
