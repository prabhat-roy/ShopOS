use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::Type, PartialEq)]
#[sqlx(type_name = "text", rename_all = "lowercase")]
#[serde(rename_all = "lowercase")]
pub enum ShipmentStatus {
    Pending,
    Picked,
    InTransit,
    Delivered,
    Failed,
    Returned,
}

impl std::fmt::Display for ShipmentStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            ShipmentStatus::Pending => write!(f, "pending"),
            ShipmentStatus::Picked => write!(f, "picked"),
            ShipmentStatus::InTransit => write!(f, "intransit"),
            ShipmentStatus::Delivered => write!(f, "delivered"),
            ShipmentStatus::Failed => write!(f, "failed"),
            ShipmentStatus::Returned => write!(f, "returned"),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct Shipment {
    pub id: Uuid,
    pub order_id: String,
    pub customer_id: String,
    pub carrier: String,
    pub tracking_number: String,
    pub status: ShipmentStatus,
    pub origin_address: serde_json::Value,
    pub dest_address: serde_json::Value,
    pub estimated_delivery: Option<DateTime<Utc>>,
    pub shipped_at: Option<DateTime<Utc>>,
    pub delivered_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateShipmentRequest {
    pub order_id: String,
    pub customer_id: String,
    pub carrier: String,
    pub origin_address: serde_json::Value,
    pub dest_address: serde_json::Value,
    pub estimated_delivery: Option<DateTime<Utc>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateStatusRequest {
    pub status: ShipmentStatus,
    pub tracking_number: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ListParams {
    pub limit: Option<i64>,
    pub offset: Option<i64>,
}
