use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant};

use redis::AsyncCommands;
use tokio::sync::RwLock;

pub struct RevocationChecker {
    client: redis::Client,
    cache: RwLock<HashMap<String, Instant>>,
    cache_ttl: Duration,
    key_prefix: String,
}

impl RevocationChecker {
    pub fn new(client: redis::Client) -> Arc<Self> {
        Arc::new(Self {
            client,
            cache: RwLock::new(HashMap::new()),
            cache_ttl: Duration::from_secs(60),
            key_prefix: "revoked:".to_string(),
        })
    }

    pub async fn is_revoked(&self, jti: &str) -> Result<bool, redis::RedisError> {
        if jti.is_empty() {
            return Ok(false);
        }

        {
            let guard = self.cache.read().await;
            if let Some(exp) = guard.get(jti) {
                if *exp > Instant::now() {
                    return Ok(true);
                }
            }
        }

        let mut conn = self.client.get_multiplexed_async_connection().await?;
        let key = format!("{}{}", self.key_prefix, jti);
        let exists: bool = conn.exists(&key).await?;

        if exists {
            let mut guard = self.cache.write().await;
            guard.insert(jti.to_string(), Instant::now() + self.cache_ttl);
        }
        Ok(exists)
    }
}
