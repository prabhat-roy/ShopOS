use async_trait::async_trait;
use sqlx::PgPool;
use uuid::Uuid;

use crate::domain::shipment::{CreateShipmentRequest, Shipment, ShipmentStatus, UpdateStatusRequest};
use crate::error::{AppError, Result};

#[async_trait]
pub trait ShipmentStore: Send + Sync {
    async fn create(&self, req: CreateShipmentRequest, tracking_number: String) -> Result<Shipment>;
    async fn get_by_id(&self, id: Uuid) -> Result<Option<Shipment>>;
    async fn get_by_order(&self, order_id: &str) -> Result<Vec<Shipment>>;
    async fn update_status(&self, id: Uuid, req: UpdateStatusRequest) -> Result<Shipment>;
    async fn list(&self, limit: i64, offset: i64) -> Result<Vec<Shipment>>;
}

pub struct PgShipmentStore {
    pool: PgPool,
}

impl PgShipmentStore {
    pub fn new(pool: PgPool) -> Self {
        Self { pool }
    }
}

#[async_trait]
impl ShipmentStore for PgShipmentStore {
    async fn create(&self, req: CreateShipmentRequest, tracking_number: String) -> Result<Shipment> {
        let status_str = "pending";
        let carrier = if req.carrier.is_empty() { "fedex".to_string() } else { req.carrier };

        let shipment = sqlx::query_as!(
            Shipment,
            r#"
            INSERT INTO shipments
                (order_id, customer_id, carrier, tracking_number, status,
                 origin_address, dest_address, estimated_delivery)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
            RETURNING
                id,
                order_id,
                customer_id,
                carrier,
                tracking_number,
                status AS "status: ShipmentStatus",
                origin_address,
                dest_address,
                estimated_delivery,
                shipped_at,
                delivered_at,
                created_at,
                updated_at
            "#,
            req.order_id,
            req.customer_id,
            carrier,
            tracking_number,
            status_str,
            req.origin_address,
            req.dest_address,
            req.estimated_delivery,
        )
        .fetch_one(&self.pool)
        .await?;

        Ok(shipment)
    }

    async fn get_by_id(&self, id: Uuid) -> Result<Option<Shipment>> {
        let shipment = sqlx::query_as!(
            Shipment,
            r#"
            SELECT
                id,
                order_id,
                customer_id,
                carrier,
                tracking_number,
                status AS "status: ShipmentStatus",
                origin_address,
                dest_address,
                estimated_delivery,
                shipped_at,
                delivered_at,
                created_at,
                updated_at
            FROM shipments
            WHERE id = $1
            "#,
            id
        )
        .fetch_optional(&self.pool)
        .await?;

        Ok(shipment)
    }

    async fn get_by_order(&self, order_id: &str) -> Result<Vec<Shipment>> {
        let shipments = sqlx::query_as!(
            Shipment,
            r#"
            SELECT
                id,
                order_id,
                customer_id,
                carrier,
                tracking_number,
                status AS "status: ShipmentStatus",
                origin_address,
                dest_address,
                estimated_delivery,
                shipped_at,
                delivered_at,
                created_at,
                updated_at
            FROM shipments
            WHERE order_id = $1
            ORDER BY created_at DESC
            "#,
            order_id
        )
        .fetch_all(&self.pool)
        .await?;

        Ok(shipments)
    }

    async fn update_status(&self, id: Uuid, req: UpdateStatusRequest) -> Result<Shipment> {
        let status_str = match req.status {
            ShipmentStatus::Pending => "pending",
            ShipmentStatus::Picked => "picked",
            ShipmentStatus::InTransit => "intransit",
            ShipmentStatus::Delivered => "delivered",
            ShipmentStatus::Failed => "failed",
            ShipmentStatus::Returned => "returned",
        };

        // Determine timestamp updates based on new status
        let shipment = sqlx::query_as!(
            Shipment,
            r#"
            UPDATE shipments SET
                status = $2,
                tracking_number = COALESCE($3, tracking_number),
                shipped_at = CASE WHEN $2 = 'intransit' AND shipped_at IS NULL
                                  THEN NOW() ELSE shipped_at END,
                delivered_at = CASE WHEN $2 = 'delivered' AND delivered_at IS NULL
                                    THEN NOW() ELSE delivered_at END,
                updated_at = NOW()
            WHERE id = $1
            RETURNING
                id,
                order_id,
                customer_id,
                carrier,
                tracking_number,
                status AS "status: ShipmentStatus",
                origin_address,
                dest_address,
                estimated_delivery,
                shipped_at,
                delivered_at,
                created_at,
                updated_at
            "#,
            id,
            status_str,
            req.tracking_number,
        )
        .fetch_optional(&self.pool)
        .await?
        .ok_or_else(|| AppError::NotFound(format!("Shipment {} not found", id)))?;

        Ok(shipment)
    }

    async fn list(&self, limit: i64, offset: i64) -> Result<Vec<Shipment>> {
        let shipments = sqlx::query_as!(
            Shipment,
            r#"
            SELECT
                id,
                order_id,
                customer_id,
                carrier,
                tracking_number,
                status AS "status: ShipmentStatus",
                origin_address,
                dest_address,
                estimated_delivery,
                shipped_at,
                delivered_at,
                created_at,
                updated_at
            FROM shipments
            ORDER BY created_at DESC
            LIMIT $1 OFFSET $2
            "#,
            limit,
            offset
        )
        .fetch_all(&self.pool)
        .await?;

        Ok(shipments)
    }
}
