use std::sync::Arc;

use anyhow::Context;
use async_trait::async_trait;
use chrono::{Duration, Utc};
use jsonwebtoken::{decode, encode, Algorithm, DecodingKey, EncodingKey, Header, Validation};
use sha2::{Digest, Sha256};
use uuid::Uuid;

use crate::{
    config::AppConfig,
    domain::auth::{Claims, Credentials, RefreshRequest, TokenPair, ValidateRequest, ValidateResponse},
    error::AuthError,
    store::auth_store::AuthStore,
};

// ---------------------------------------------------------------------------
// AuthService trait
// ---------------------------------------------------------------------------

#[async_trait]
pub trait AuthService: Send + Sync {
    async fn login(&self, creds: Credentials) -> Result<TokenPair, AuthError>;
    async fn refresh(&self, req: RefreshRequest) -> Result<TokenPair, AuthError>;
    async fn validate(&self, req: ValidateRequest) -> Result<ValidateResponse, AuthError>;
    async fn logout(&self, refresh_token: String) -> Result<(), AuthError>;
}

// ---------------------------------------------------------------------------
// Concrete implementation
// ---------------------------------------------------------------------------

pub struct AuthServiceImpl {
    store: Arc<dyn AuthStore>,
    config: AppConfig,
}

impl AuthServiceImpl {
    pub fn new(store: Arc<dyn AuthStore>, config: AppConfig) -> Self {
        Self { store, config }
    }

    // ------------------------------------------------------------------
    // Private helpers
    // ------------------------------------------------------------------

    /// Sign a new JWT access token for `user_id` / `email`.
    fn mint_access_token(&self, user_id: Uuid, email: &str) -> Result<String, AuthError> {
        let now = Utc::now();
        let exp = (now + Duration::seconds(self.config.jwt_expiry_seconds)).timestamp();

        let claims = Claims {
            sub: user_id.to_string(),
            email: email.to_owned(),
            exp,
            iat: now.timestamp(),
        };

        let token = encode(
            &Header::new(Algorithm::HS256),
            &claims,
            &EncodingKey::from_secret(self.config.jwt_secret.as_bytes()),
        )?;

        Ok(token)
    }

    /// Generate a high-entropy opaque refresh token (two UUID v4s concatenated).
    fn generate_refresh_token() -> String {
        format!("{}{}", Uuid::new_v4().simple(), Uuid::new_v4().simple())
    }

    /// SHA-256 hash of a raw refresh token, hex-encoded, for safe storage.
    fn hash_token(token: &str) -> String {
        let mut hasher = Sha256::new();
        hasher.update(token.as_bytes());
        hex::encode(hasher.finalize())
    }

    /// Decode and verify a JWT access token, returning its embedded claims.
    fn decode_access_token(&self, token: &str) -> Result<Claims, AuthError> {
        let mut validation = Validation::new(Algorithm::HS256);
        validation.validate_exp = true;

        let data = decode::<Claims>(
            token,
            &DecodingKey::from_secret(self.config.jwt_secret.as_bytes()),
            &validation,
        )?;

        Ok(data.claims)
    }

    /// Build a full `TokenPair` and persist the refresh token.
    async fn issue_token_pair(
        &self,
        user_id: Uuid,
        email: &str,
    ) -> Result<TokenPair, AuthError> {
        let access_token = self.mint_access_token(user_id, email)?;

        let refresh_token = Self::generate_refresh_token();
        let token_hash = Self::hash_token(&refresh_token);
        let expires_at = Utc::now() + Duration::days(self.config.refresh_expiry_days);

        self.store
            .save_refresh_token(user_id, &token_hash, expires_at)
            .await
            .context("Failed to persist refresh token")?;

        Ok(TokenPair {
            access_token,
            refresh_token,
            expires_in: self.config.jwt_expiry_seconds,
        })
    }
}

#[async_trait]
impl AuthService for AuthServiceImpl {
    // ------------------------------------------------------------------
    // login
    // ------------------------------------------------------------------
    async fn login(&self, creds: Credentials) -> Result<TokenPair, AuthError> {
        // 1. Fetch user record
        let user = self
            .store
            .find_user_by_email(&creds.email)
            .await
            .context("find_user_by_email")?
            .ok_or(AuthError::UserNotFound)?;

        // 2. Verify bcrypt password hash on a blocking thread (CPU-bound)
        let password = creds.password.clone();
        let hash = user.password_hash.clone();
        let valid = tokio::task::spawn_blocking(move || bcrypt::verify(&password, &hash))
            .await
            .context("bcrypt thread panicked")?
            .context("bcrypt::verify error")?;

        if !valid {
            return Err(AuthError::InvalidCredentials);
        }

        // 3. Issue and return token pair
        self.issue_token_pair(user.id, &user.email).await
    }

    // ------------------------------------------------------------------
    // refresh
    // ------------------------------------------------------------------
    async fn refresh(&self, req: RefreshRequest) -> Result<TokenPair, AuthError> {
        let token_hash = Self::hash_token(&req.refresh_token);

        // 1. Locate the stored record
        let record = self
            .store
            .find_refresh_token(&token_hash)
            .await
            .context("find_refresh_token")?
            .ok_or(AuthError::RefreshTokenNotFound)?;

        // 2. Enforce expiry
        if record.expires_at < Utc::now() {
            let _ = self.store.revoke_refresh_token(&token_hash).await;
            return Err(AuthError::RefreshTokenExpired);
        }

        // 3. Resolve the owner's current email
        let user = self
            .store
            .find_user_by_id(record.user_id)
            .await
            .context("find_user_by_id")?
            .ok_or(AuthError::UserNotFound)?;

        // 4. Rotate: revoke old token, issue new pair
        self.store
            .revoke_refresh_token(&token_hash)
            .await
            .context("revoke_refresh_token")?;

        self.issue_token_pair(user.id, &user.email).await
    }

    // ------------------------------------------------------------------
    // validate
    // ------------------------------------------------------------------
    async fn validate(&self, req: ValidateRequest) -> Result<ValidateResponse, AuthError> {
        match self.decode_access_token(&req.token) {
            Ok(claims) => Ok(ValidateResponse {
                valid: true,
                user_id: claims.sub,
                email: claims.email,
            }),
            Err(AuthError::Jwt(jwt_err)) => {
                use jsonwebtoken::errors::ErrorKind;
                match jwt_err.kind() {
                    ErrorKind::ExpiredSignature => Err(AuthError::TokenExpired),
                    _ => Err(AuthError::InvalidToken),
                }
            }
            Err(e) => Err(e),
        }
    }

    // ------------------------------------------------------------------
    // logout
    // ------------------------------------------------------------------
    async fn logout(&self, refresh_token: String) -> Result<(), AuthError> {
        let token_hash = Self::hash_token(&refresh_token);
        self.store
            .revoke_refresh_token(&token_hash)
            .await
            .context("revoke_refresh_token")?;
        Ok(())
    }
}
