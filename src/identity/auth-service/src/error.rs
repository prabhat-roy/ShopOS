use axum::{
    http::StatusCode,
    response::{IntoResponse, Response},
    Json,
};
use serde_json::json;
use thiserror::Error;

#[derive(Debug, Error)]
pub enum AuthError {
    #[error("Invalid credentials")]
    InvalidCredentials,

    #[error("User not found")]
    UserNotFound,

    #[error("Token expired")]
    TokenExpired,

    #[error("Invalid token")]
    InvalidToken,

    #[error("Refresh token not found or already revoked")]
    RefreshTokenNotFound,

    #[error("Refresh token expired")]
    RefreshTokenExpired,

    #[error("Database error: {0}")]
    Database(#[from] sqlx::Error),

    #[error("JWT error: {0}")]
    Jwt(#[from] jsonwebtoken::errors::Error),

    #[error("Internal error: {0}")]
    Internal(#[from] anyhow::Error),
}

impl IntoResponse for AuthError {
    fn into_response(self) -> Response {
        let (status, message): (StatusCode, String) = match &self {
            AuthError::InvalidCredentials => {
                (StatusCode::UNAUTHORIZED, "Invalid credentials".into())
            }
            AuthError::UserNotFound => {
                // Mask existence: return same message as bad password
                (StatusCode::UNAUTHORIZED, "Invalid credentials".into())
            }
            AuthError::TokenExpired => (StatusCode::UNAUTHORIZED, "Token expired".into()),
            AuthError::InvalidToken => (StatusCode::UNAUTHORIZED, "Invalid token".into()),
            AuthError::RefreshTokenNotFound => {
                (StatusCode::UNAUTHORIZED, "Refresh token not found or already revoked".into())
            }
            AuthError::RefreshTokenExpired => {
                (StatusCode::UNAUTHORIZED, "Refresh token expired".into())
            }
            AuthError::Database(e) => {
                tracing::error!(error = %e, "Database error");
                (StatusCode::INTERNAL_SERVER_ERROR, "Internal server error".into())
            }
            AuthError::Jwt(e) => {
                tracing::error!(error = %e, "JWT error");
                (StatusCode::UNAUTHORIZED, "Token error".into())
            }
            AuthError::Internal(e) => {
                tracing::error!(error = %e, "Internal error");
                (StatusCode::INTERNAL_SERVER_ERROR, "Internal server error".into())
            }
        };

        let body = Json(json!({ "error": message }));
        (status, body).into_response()
    }
}
