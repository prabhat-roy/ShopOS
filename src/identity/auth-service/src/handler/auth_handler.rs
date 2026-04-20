use axum::{
    extract::State,
    http::StatusCode,
    response::IntoResponse,
    Json,
};
use serde_json::json;
use std::sync::Arc;

use crate::{
    domain::auth::{Credentials, LogoutRequest, RefreshRequest, ValidateRequest},
    error::AuthError,
    service::auth_service::{AuthService as _},
    AppState,
};

// ---------------------------------------------------------------------------
// GET /healthz
// ---------------------------------------------------------------------------

pub async fn healthz() -> impl IntoResponse {
    (StatusCode::OK, Json(json!({"status": "ok"})))
}

// ---------------------------------------------------------------------------
// POST /auth/login
//
// Body: { "email": "...", "password": "..." }
// Response 200: TokenPair
// ---------------------------------------------------------------------------

pub async fn login(
    State(state): State<Arc<AppState>>,
    Json(creds): Json<Credentials>,
) -> Result<impl IntoResponse, AuthError> {
    let token_pair = state.auth_service.login(creds).await?;
    Ok((StatusCode::OK, Json(token_pair)))
}

// ---------------------------------------------------------------------------
// POST /auth/refresh
//
// Body: { "refresh_token": "..." }
// Response 200: TokenPair
// ---------------------------------------------------------------------------

pub async fn refresh(
    State(state): State<Arc<AppState>>,
    Json(req): Json<RefreshRequest>,
) -> Result<impl IntoResponse, AuthError> {
    let token_pair = state.auth_service.refresh(req).await?;
    Ok((StatusCode::OK, Json(token_pair)))
}

// ---------------------------------------------------------------------------
// POST /auth/logout
//
// Body: { "refresh_token": "..." }
// Response 204: No Content
// ---------------------------------------------------------------------------

pub async fn logout(
    State(state): State<Arc<AppState>>,
    Json(req): Json<LogoutRequest>,
) -> Result<impl IntoResponse, AuthError> {
    state.auth_service.logout(req.refresh_token).await?;
    Ok(StatusCode::NO_CONTENT)
}

// ---------------------------------------------------------------------------
// POST /auth/validate
//
// Body: { "token": "..." }
// Response 200: ValidateResponse
// ---------------------------------------------------------------------------

pub async fn validate(
    State(state): State<Arc<AppState>>,
    Json(req): Json<ValidateRequest>,
) -> Result<impl IntoResponse, AuthError> {
    let resp = state.auth_service.validate(req).await?;
    Ok((StatusCode::OK, Json(resp)))
}
