use std::sync::Arc;

use uuid::Uuid;

use crate::domain::shipment::{CreateShipmentRequest, Shipment, UpdateStatusRequest};
use crate::error::Result;
use crate::store::shipment_store::ShipmentStore;

pub struct ShippingService {
    store: Arc<dyn ShipmentStore>,
}

impl ShippingService {
    pub fn new(store: Arc<dyn ShipmentStore>) -> Self {
        Self { store }
    }

    /// Generate a short tracking number of the form TRACK-XXXXXXXX
    fn generate_tracking_number() -> String {
        let id = Uuid::new_v4();
        let short = &id.to_string().replace('-', "")[..8].to_uppercase();
        format!("TRACK-{}", short)
    }

    pub async fn create(&self, req: CreateShipmentRequest) -> Result<Shipment> {
        let tracking = Self::generate_tracking_number();
        tracing::debug!("Creating shipment for order {} with tracking {}", req.order_id, tracking);
        self.store.create(req, tracking).await
    }

    pub async fn get(&self, id: Uuid) -> Result<Option<Shipment>> {
        self.store.get_by_id(id).await
    }

    pub async fn get_by_order(&self, order_id: &str) -> Result<Vec<Shipment>> {
        self.store.get_by_order(order_id).await
    }

    pub async fn update_status(&self, id: Uuid, req: UpdateStatusRequest) -> Result<Shipment> {
        tracing::debug!("Updating shipment {} status to {:?}", id, req.status);
        self.store.update_status(id, req).await
    }

    pub async fn list(&self, limit: i64, offset: i64) -> Result<Vec<Shipment>> {
        let limit = limit.min(100).max(1);
        let offset = offset.max(0);
        self.store.list(limit, offset).await
    }
}
