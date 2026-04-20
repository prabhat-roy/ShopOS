use anyhow::Context;

#[derive(Debug, Clone)]
pub struct Config {
    pub http_port: u16,
    pub grpc_port: u16,
    pub database_url: String,
    pub rust_log: String,
}

impl Config {
    pub fn from_env() -> anyhow::Result<Self> {
        Ok(Self {
            http_port: std::env::var("HTTP_PORT")
                .unwrap_or_else(|_| "8138".to_string())
                .parse::<u16>()
                .context("HTTP_PORT must be a valid port number")?,
            grpc_port: std::env::var("GRPC_PORT")
                .unwrap_or_else(|_| "50084".to_string())
                .parse::<u16>()
                .context("GRPC_PORT must be a valid port number")?,
            database_url: std::env::var("DATABASE_URL")
                .context("DATABASE_URL must be set")?,
            rust_log: std::env::var("RUST_LOG").unwrap_or_else(|_| "info".to_string()),
        })
    }
}
