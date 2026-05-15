use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum Role {
    Customer,
    Barber,
    ShopOwner,
    Admin,
}

#[derive(Debug, Clone)]
pub struct User {
    pub id: String,
    pub role: Role,
    pub shop_id: Option<String>,
    pub jti: Option<String>,
}

#[derive(Debug, Deserialize)]
pub(crate) struct Claims {
    pub sub: String,
    pub role: Role,
    #[serde(default)]
    pub shop_id: Option<String>,
    #[serde(default)]
    pub jti: Option<String>,
    pub exp: usize,
    pub iat: usize,
    pub iss: String,
    pub aud: String,
}
