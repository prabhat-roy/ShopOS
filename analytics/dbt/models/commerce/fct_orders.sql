-- Fact table: one row per order with enriched metrics
{{
  config(
    materialized='incremental',
    unique_key='order_id',
    incremental_strategy='merge',
    on_schema_change='append_new_columns'
  )
}}

WITH orders AS (
    SELECT * FROM {{ ref('stg_orders') }}
    {% if is_incremental() %}
        WHERE ordered_at > (SELECT MAX(ordered_at) FROM {{ this }})
    {% endif %}
),

order_items_agg AS (
    SELECT
        order_id,
        COUNT(*)                                AS item_count,
        SUM(quantity)                           AS total_units,
        SUM(discount_amount_usd)                AS total_discount_usd,
        COUNT(DISTINCT product_id)              AS unique_products,
        COUNT(DISTINCT category_id)             AS unique_categories
    FROM {{ ref('stg_order_items') }}
    GROUP BY 1
),

final AS (
    SELECT
        o.order_id,
        o.user_id,
        o.status,
        o.status_key,
        o.total_usd,
        o.subtotal_usd,
        o.shipping_usd,
        o.tax_usd,
        o.currency,
        o.coupon_code,
        o.shipping_country,
        o.payment_method,
        o.ordered_at,
        o.ordered_date,
        o.ordered_week,
        o.ordered_month,
        o.ordered_hour,
        COALESCE(i.item_count, 0)               AS item_count,
        COALESCE(i.total_units, 0)              AS total_units,
        COALESCE(i.total_discount_usd, 0)       AS total_discount_usd,
        COALESCE(i.unique_products, 0)          AS unique_products,
        COALESCE(i.unique_categories, 0)        AS unique_categories,
        -- Derived
        o.status_key IN ('delivered', 'completed') AS is_completed,
        o.status_key IN ('cancelled', 'refunded')  AS is_cancelled,
        o.coupon_code IS NOT NULL                   AS has_coupon
    FROM orders o
    LEFT JOIN order_items_agg i USING (order_id)
)

SELECT * FROM final
