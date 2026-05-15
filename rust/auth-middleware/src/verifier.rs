use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant};

use jsonwebtoken::{decode, decode_header, DecodingKey, Validation};
use serde::Deserialize;
use thiserror::Error;
use tokio::sync::RwLock;

use crate::context::{Claims, Role, User};

#[derive(Debug, Error)]
pub enum VerifyError {
    #[error("missing or malformed authorization header")]
    BadHeader,
    #[error("unknown key id")]
    UnknownKid,
    #[error("token invalid: {0}")]
    Invalid(#[from] jsonwebtoken::errors::Error),
    #[error("jwks fetch failed: {0}")]
    JwksFetch(#[from] reqwest::Error),
}

#[derive(Debug, Deserialize)]
struct Jwks {
    keys: Vec<Jwk>,
}

#[derive(Debug, Deserialize)]
struct Jwk {
    kid: String,
    kty: String,
    #[serde(default)]
    n: Option<String>,
    #[serde(default)]
    e: Option<String>,
    #[serde(default, rename = "use")]
    use_: Option<String>,
    #[serde(default)]
    alg: Option<String>,
    // Ed25519 keys use 'x' instead of n/e.
    #[serde(default)]
    x: Option<String>,
    #[serde(default)]
    crv: Option<String>,
}

struct CacheEntry {
    keys: HashMap<String, DecodingKey>,
    fetched_at: Instant,
}

pub struct Verifier {
    jwks_url: String,
    issuer: String,
    audience: String,
    cache_ttl: Duration,
    http: reqwest::Client,
    cache: RwLock<Option<CacheEntry>>,
}

impl Verifier {
    pub fn new(jwks_url: impl Into<String>, issuer: impl Into<String>, audience: impl Into<String>) -> Arc<Self> {
        Arc::new(Self {
            jwks_url: jwks_url.into(),
            issuer: issuer.into(),
            audience: audience.into(),
            cache_ttl: Duration::from_secs(600),
            http: reqwest::Client::new(),
            cache: RwLock::new(None),
        })
    }

    async fn refresh(&self) -> Result<(), VerifyError> {
        let jwks: Jwks = self.http.get(&self.jwks_url).send().await?.error_for_status()?.json().await?;
        let mut map = HashMap::new();
        for k in jwks.keys {
            let key = match k.kty.as_str() {
                "RSA" => {
                    let n = k.n.unwrap_or_default();
                    let e = k.e.unwrap_or_default();
                    match DecodingKey::from_rsa_components(&n, &e) {
                        Ok(v) => v,
                        Err(_) => continue,
                    }
                }
                "OKP" => {
                    let x = k.x.unwrap_or_default();
                    match DecodingKey::from_ed_components(&x) {
                        Ok(v) => v,
                        Err(_) => continue,
                    }
                }
                _ => continue,
            };
            map.insert(k.kid, key);
        }
        let mut guard = self.cache.write().await;
        *guard = Some(CacheEntry { keys: map, fetched_at: Instant::now() });
        Ok(())
    }

    pub async fn refresh_now(&self) -> Result<(), VerifyError> {
        self.refresh().await
    }

    pub async fn verify(&self, token: &str) -> Result<User, VerifyError> {
        let stale = {
            let guard = self.cache.read().await;
            match guard.as_ref() {
                Some(c) => c.fetched_at.elapsed() > self.cache_ttl,
                None => true,
            }
        };
        if stale {
            self.refresh().await?;
        }

        let header = decode_header(token)?;
        let kid = header.kid.ok_or(VerifyError::UnknownKid)?;

        let key = {
            let guard = self.cache.read().await;
            let entry = guard.as_ref().ok_or(VerifyError::UnknownKid)?;
            entry.keys.get(&kid).cloned().ok_or(VerifyError::UnknownKid)?
        };

        let mut validation = Validation::new(header.alg);
        validation.set_issuer(&[&self.issuer]);
        validation.set_audience(&[&self.audience]);

        let data = decode::<Claims>(token, &key, &validation)?;
        Ok(User {
            id: data.claims.sub,
            role: data.claims.role,
            shop_id: data.claims.shop_id,
            jti: data.claims.jti,
        })
    }
}

#[allow(dead_code)]
fn _role_kind(_: Role) {}
