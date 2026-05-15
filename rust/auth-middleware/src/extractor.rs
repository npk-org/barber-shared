use axum::extract::FromRequestParts;
use axum::http::request::Parts;
use axum::http::StatusCode;

use crate::context::User;

#[async_trait::async_trait]
impl<S> FromRequestParts<S> for User
where
    S: Send + Sync,
{
    type Rejection = (StatusCode, &'static str);

    async fn from_request_parts(parts: &mut Parts, _state: &S) -> Result<Self, Self::Rejection> {
        parts
            .extensions
            .get::<User>()
            .cloned()
            .ok_or((StatusCode::UNAUTHORIZED, "no user in request extensions"))
    }
}
