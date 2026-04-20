mod config;
mod domain;
mod error;
mod handler;
mod service;
mod store;

use axum::{
    routing::{get, post},
    Router,
};
use sqlx::postgres::PgPoolOptions;
use std::sync::Arc;
use tower_http::{cors::CorsLayer, trace::TraceLayer};
use tracing::info;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt, EnvFilter};

use crate::{
    config::AppConfig,
    handler::auth_handler::{healthz, login, logout, refresh, validate},
    service::auth_service::AuthServiceImpl,
    store::auth_store::PgAuthStore,
};

pub struct AppState {
    pub auth_service: Arc<AuthServiceImpl>,
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    // Load .env file if present (non-fatal if missing)
    let _ = dotenvy::dotenv();

    // Initialise tracing
    tracing_subscriber::registry()
        .with(EnvFilter::try_from_default_env().unwrap_or_else(|_| "info".into()))
        .with(tracing_subscriber::fmt::layer())
        .init();

    let cfg = AppConfig::from_env()?;
    info!(
        http_port = cfg.http_port,
        grpc_port = cfg.grpc_port,
        "Starting shopos-auth-service"
    );

    // Database pool (lazy — connects on first query, not at startup)
    let pool = PgPoolOptions::new()
        .max_connections(20)
        .connect_lazy(&cfg.database_url)?;
    info!("Postgres pool created (lazy connection)");

    let store = Arc::new(PgAuthStore::new(pool));
    let auth_service = Arc::new(AuthServiceImpl::new(store, cfg.clone()));
    let state = Arc::new(AppState { auth_service });

    let app = Router::new()
        .route("/healthz", get(healthz))
        .route("/auth/login", post(login))
        .route("/auth/refresh", post(refresh))
        .route("/auth/logout", post(logout))
        .route("/auth/validate", post(validate))
        .layer(CorsLayer::permissive())
        .layer(TraceLayer::new_for_http())
        .with_state(state);

    let addr = format!("0.0.0.0:{}", cfg.http_port);
    let listener = tokio::net::TcpListener::bind(&addr).await?;
    info!(address = %addr, "HTTP server listening");

    axum::serve(listener, app).await?;
    Ok(())
}
