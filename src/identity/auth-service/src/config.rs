use anyhow::{Context, Result};

/// All runtime configuration sourced from environment variables.
#[derive(Debug, Clone)]
pub struct AppConfig {
    pub http_port: u16,
    pub grpc_port: u16,
    pub database_url: String,
    pub jwt_secret: String,
    /// Access token lifetime in seconds (default 900 = 15 min).
    pub jwt_expiry_seconds: i64,
    /// Refresh token lifetime in days (default 7).
    pub refresh_expiry_days: i64,
}

impl AppConfig {
    pub fn from_env() -> Result<Self> {
        Ok(Self {
            http_port: env_var("HTTP_PORT")
                .unwrap_or_else(|_| "8098".into())
                .parse::<u16>()
                .context("HTTP_PORT must be a valid port number")?,

            grpc_port: env_var("GRPC_PORT")
                .unwrap_or_else(|_| "50060".into())
                .parse::<u16>()
                .context("GRPC_PORT must be a valid port number")?,

            database_url: env_var("DATABASE_URL")
                .context("DATABASE_URL is required")?,

            jwt_secret: env_var("JWT_SECRET")
                .context("JWT_SECRET is required")?,

            jwt_expiry_seconds: env_var("JWT_EXPIRY_SECONDS")
                .unwrap_or_else(|_| "900".into())
                .parse::<i64>()
                .context("JWT_EXPIRY_SECONDS must be an integer")?,

            refresh_expiry_days: env_var("REFRESH_EXPIRY_DAYS")
                .unwrap_or_else(|_| "7".into())
                .parse::<i64>()
                .context("REFRESH_EXPIRY_DAYS must be an integer")?,
        })
    }
}

fn env_var(key: &str) -> Result<String> {
    std::env::var(key).with_context(|| format!("Missing env var: {key}"))
}
