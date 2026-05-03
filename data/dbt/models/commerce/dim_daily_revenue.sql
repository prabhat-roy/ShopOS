-- Daily revenue summary — used by Grafana dashboards and reporting-service
{{
  config(
    materialized='incremental',
    unique_key='ordered_date',
    incremental_strategy='merge'
  )
}}

WITH base AS (
    SELECT * FROM {{ ref('fct_orders') }}
    WHERE NOT is_cancelled
    {% if is_incremental() %}
        AND ordered_date >= (SELECT MAX(ordered_date) - INTERVAL '3 days' FROM {{ this }})
    {% endif %}
)

SELECT
    ordered_date,
    ordered_month,
    ordered_week,
    COUNT(*)                                            AS order_count,
    COUNT(DISTINCT user_id)                             AS unique_buyers,
    SUM(total_usd)                                      AS gross_revenue,
    SUM(total_discount_usd)                             AS total_discounts,
    SUM(total_usd) - SUM(total_discount_usd)            AS net_revenue,
    SUM(shipping_usd)                                   AS shipping_revenue,
    SUM(tax_usd)                                        AS tax_collected,
    AVG(total_usd)                                      AS avg_order_value,
    PERCENTILE_CONT(0.5)  WITHIN GROUP (ORDER BY total_usd) AS median_order_value,
    SUM(item_count)                                     AS total_items_sold,
    SUM(total_units)                                    AS total_units_sold,
    COUNT(*) FILTER (WHERE has_coupon)                  AS orders_with_coupon,
    AVG(total_usd) FILTER (WHERE has_coupon)            AS avg_order_value_with_coupon
FROM base
GROUP BY ordered_date, ordered_month, ordered_week
ORDER BY ordered_date
