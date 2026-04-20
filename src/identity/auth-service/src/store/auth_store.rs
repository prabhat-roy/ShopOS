use anyhow::Result;
use async_trait::async_trait;
use chrono::{DateTime, Utc};
use sqlx::PgPool;
use uuid::Uuid;

// ---------------------------------------------------------------------------
// Record types
// ---------------------------------------------------------------------------

/// A row from the `users` table.
#[derive(Debug, sqlx::FromRow)]
pub struct UserRecord {
    pub id: Uuid,
    pub email: String,
    pub password_hash: String,
}

/// A row from the `refresh_tokens` table.
#[derive(Debug, sqlx::FromRow)]
pub struct RefreshTokenRecord {
    pub id: Uuid,
    pub user_id: Uuid,
    pub token_hash: String,
    pub expires_at: DateTime<Utc>,
}

// ---------------------------------------------------------------------------
// Trait
// ---------------------------------------------------------------------------

#[async_trait]
pub trait AuthStore: Send + Sync {
    /// Look up a user by e-mail address.
    async fn find_user_by_email(&self, email: &str) -> Result<Option<UserRecord>>;

    /// Look up a user by UUID (used during token refresh).
    async fn find_user_by_id(&self, id: Uuid) -> Result<Option<UserRecord>>;

    /// Persist a hashed refresh token for the given user.
    async fn save_refresh_token(
        &self,
        user_id: Uuid,
        token_hash: &str,
        expires_at: DateTime<Utc>,
    ) -> Result<()>;

    /// Retrieve a refresh token record by its hash.
    async fn find_refresh_token(&self, token_hash: &str) -> Result<Option<RefreshTokenRecord>>;

    /// Delete a single refresh token (logout / rotation).
    async fn revoke_refresh_token(&self, token_hash: &str) -> Result<()>;

    /// Delete all refresh tokens that belong to a user (forced logout everywhere).
    async fn revoke_all_user_tokens(&self, user_id: Uuid) -> Result<()>;
}

// ---------------------------------------------------------------------------
// Postgres implementation
// ---------------------------------------------------------------------------

pub struct PgAuthStore {
    pool: PgPool,
}

impl PgAuthStore {
    pub fn new(pool: PgPool) -> Self {
        Self { pool }
    }
}

#[async_trait]
impl AuthStore for PgAuthStore {
    async fn find_user_by_email(&self, email: &str) -> Result<Option<UserRecord>> {
        let record = sqlx::query_as!(
            UserRecord,
            r#"
            SELECT id, email, password_hash
            FROM users
            WHERE email = $1
            "#,
            email
        )
        .fetch_optional(&self.pool)
        .await?;

        Ok(record)
    }

    async fn find_user_by_id(&self, id: Uuid) -> Result<Option<UserRecord>> {
        let record = sqlx::query_as!(
            UserRecord,
            r#"
            SELECT id, email, password_hash
            FROM users
            WHERE id = $1
            "#,
            id
        )
        .fetch_optional(&self.pool)
        .await?;

        Ok(record)
    }

    async fn save_refresh_token(
        &self,
        user_id: Uuid,
        token_hash: &str,
        expires_at: DateTime<Utc>,
    ) -> Result<()> {
        sqlx::query!(
            r#"
            INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
            VALUES ($1, $2, $3)
            "#,
            user_id,
            token_hash,
            expires_at
        )
        .execute(&self.pool)
        .await?;

        Ok(())
    }

    async fn find_refresh_token(&self, token_hash: &str) -> Result<Option<RefreshTokenRecord>> {
        let record = sqlx::query_as!(
            RefreshTokenRecord,
            r#"
            SELECT id, user_id, token_hash, expires_at
            FROM refresh_tokens
            WHERE token_hash = $1
            "#,
            token_hash
        )
        .fetch_optional(&self.pool)
        .await?;

        Ok(record)
    }

    async fn revoke_refresh_token(&self, token_hash: &str) -> Result<()> {
        sqlx::query!(
            r#"
            DELETE FROM refresh_tokens
            WHERE token_hash = $1
            "#,
            token_hash
        )
        .execute(&self.pool)
        .await?;

        Ok(())
    }

    async fn revoke_all_user_tokens(&self, user_id: Uuid) -> Result<()> {
        sqlx::query!(
            r#"
            DELETE FROM refresh_tokens
            WHERE user_id = $1
            "#,
            user_id
        )
        .execute(&self.pool)
        .await?;

        Ok(())
    }
}
