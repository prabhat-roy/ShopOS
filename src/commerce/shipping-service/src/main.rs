mod config;
mod domain;
mod error;
mod handler;
mod service;
mod store;

use std::sync::Arc;

use axum::Router;
use sqlx::postgres::PgPoolOptions;
use tracing::info;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use crate::config::Config;
use crate::handler::shipping_handler::build_router;
use crate::service::shipping_service::ShippingService;
use crate::store::shipment_store::PgShipmentStore;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    dotenvy::dotenv().ok();

    tracing_subscriber::registry()
        .with(tracing_subscriber::EnvFilter::try_from_default_env().unwrap_or_else(|_| {
            "shipping_service=debug,tower_http=debug".into()
        }))
        .with(tracing_subscriber::fmt::layer())
        .init();

    let cfg = Config::from_env()?;
    info!("Starting shipping-service on port {}", cfg.http_port);

    let pool = PgPoolOptions::new()
        .max_connections(10)
        .connect_lazy(&cfg.database_url)?;

    info!("Postgres pool created (lazy connection)");

    if let Err(e) = sqlx::migrate!("./db/migrations").run(&pool).await {
        tracing::warn!("Migrations skipped — DB unavailable: {}", e);
    } else {
        info!("Migrations applied");
    }

    let store = Arc::new(PgShipmentStore::new(pool));
    let service = Arc::new(ShippingService::new(store));
    let app: Router = build_router(service);

    let addr = format!("0.0.0.0:{}", cfg.http_port);
    let listener = tokio::net::TcpListener::bind(&addr).await?;
    info!("Listening on {}", addr);

    axum::serve(listener, app).await?;
    Ok(())
}
