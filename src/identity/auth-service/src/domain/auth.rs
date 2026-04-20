use serde::{Deserialize, Serialize};

/// Credentials submitted by a user during login.
#[derive(Debug, Deserialize)]
pub struct Credentials {
    pub email: String,
    pub password: String,
}

/// Pair of tokens returned after a successful login or refresh.
#[derive(Debug, Serialize)]
pub struct TokenPair {
    pub access_token: String,
    pub refresh_token: String,
    /// Access token lifetime in seconds.
    pub expires_in: i64,
}

/// JWT claims embedded in every access token.
#[derive(Debug, Serialize, Deserialize)]
pub struct Claims {
    /// Subject — the user UUID as a string.
    pub sub: String,
    pub email: String,
    /// Expiry (Unix timestamp).
    pub exp: i64,
    /// Issued-at (Unix timestamp).
    pub iat: i64,
}

/// Body for the refresh endpoint.
#[derive(Debug, Deserialize)]
pub struct RefreshRequest {
    pub refresh_token: String,
}

/// Body for the validate endpoint.
#[derive(Debug, Deserialize)]
pub struct ValidateRequest {
    pub token: String,
}

/// Response from the validate endpoint.
#[derive(Debug, Serialize)]
pub struct ValidateResponse {
    pub valid: bool,
    pub user_id: String,
    pub email: String,
}

/// Body for the logout endpoint.
#[derive(Debug, Deserialize)]
pub struct LogoutRequest {
    pub refresh_token: String,
}
