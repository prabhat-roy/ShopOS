-- Product performance fact table — units sold, revenue, returns per product
{{
  config(
    materialized='incremental',
    unique_key=['product_id', 'ordered_date'],
    incremental_strategy='merge'
  )
}}

WITH items AS (
    SELECT
        oi.product_id,
        oi.sku,
        oi.category_id,
        oi.brand_id,
        o.ordered_date,
        o.ordered_month,
        SUM(oi.quantity)            AS units_sold,
        SUM(oi.total_price_usd)     AS revenue_usd,
        SUM(oi.discount_amount_usd) AS discount_usd,
        COUNT(DISTINCT oi.order_id) AS order_count,
        COUNT(*) FILTER (WHERE o.is_cancelled) AS cancelled_count
    FROM {{ ref('stg_order_items') }} oi
    JOIN {{ ref('fct_orders') }} o USING (order_id)
    {% if is_incremental() %}
        WHERE o.ordered_date >= (SELECT MAX(ordered_date) - INTERVAL '3 days' FROM {{ this }})
    {% endif %}
    GROUP BY 1, 2, 3, 4, 5, 6
)

SELECT
    product_id,
    sku,
    category_id,
    brand_id,
    ordered_date,
    ordered_month,
    units_sold,
    revenue_usd,
    discount_usd,
    revenue_usd - discount_usd  AS net_revenue_usd,
    order_count,
    cancelled_count,
    ROUND(100.0 * cancelled_count / NULLIF(order_count, 0), 2) AS cancellation_rate_pct
FROM items
