use std::sync::Arc;

use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, patch, post},
    Json, Router,
};
use serde::Deserialize;
use serde_json::json;
use uuid::Uuid;

use crate::domain::shipment::{CreateShipmentRequest, UpdateStatusRequest};
use crate::error::{AppError, Result};
use crate::service::shipping_service::ShippingService;

pub type ServiceState = Arc<ShippingService>;

pub fn build_router(service: Arc<ShippingService>) -> Router {
    Router::new()
        .route("/healthz", get(healthz))
        .route("/shipments", post(create_shipment).get(list_shipments))
        .route("/shipments/:id", get(get_shipment))
        .route("/shipments/:id/status", patch(update_status))
        .route("/shipments/order/:order_id", get(get_by_order))
        .with_state(service)
}

async fn healthz() -> impl IntoResponse {
    Json(json!({ "status": "ok" }))
}

async fn create_shipment(
    State(svc): State<ServiceState>,
    Json(req): Json<CreateShipmentRequest>,
) -> Result<impl IntoResponse> {
    if req.order_id.is_empty() {
        return Err(AppError::BadRequest("order_id is required".into()));
    }
    if req.customer_id.is_empty() {
        return Err(AppError::BadRequest("customer_id is required".into()));
    }

    let shipment = svc.create(req).await?;
    Ok((StatusCode::CREATED, Json(shipment)))
}

async fn get_shipment(
    State(svc): State<ServiceState>,
    Path(id): Path<Uuid>,
) -> Result<impl IntoResponse> {
    match svc.get(id).await? {
        Some(s) => Ok(Json(s)),
        None => Err(AppError::NotFound(format!("Shipment {} not found", id))),
    }
}

async fn get_by_order(
    State(svc): State<ServiceState>,
    Path(order_id): Path<String>,
) -> Result<impl IntoResponse> {
    let shipments = svc.get_by_order(&order_id).await?;
    Ok(Json(shipments))
}

async fn update_status(
    State(svc): State<ServiceState>,
    Path(id): Path<Uuid>,
    Json(req): Json<UpdateStatusRequest>,
) -> Result<impl IntoResponse> {
    let shipment = svc.update_status(id, req).await?;
    Ok(Json(shipment))
}

#[derive(Debug, Deserialize)]
struct ListQuery {
    limit: Option<i64>,
    offset: Option<i64>,
}

async fn list_shipments(
    State(svc): State<ServiceState>,
    Query(params): Query<ListQuery>,
) -> Result<impl IntoResponse> {
    let limit = params.limit.unwrap_or(20);
    let offset = params.offset.unwrap_or(0);
    let shipments = svc.list(limit, offset).await?;
    let count = shipments.len();
    Ok(Json(json!({
        "data": shipments,
        "limit": limit,
        "offset": offset,
        "count": count
    })))
}
