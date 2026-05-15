use std::sync::Arc;
use std::task::{Context, Poll};

use axum::body::Body;
use axum::http::{Request, Response, StatusCode};
use axum::response::IntoResponse;
use tower::Layer;
use tower_service::Service;

use crate::revocation::RevocationChecker;
use crate::verifier::Verifier;

#[derive(Clone)]
pub struct AuthLayer {
    pub verifier: Arc<Verifier>,
    pub revocation: Option<Arc<RevocationChecker>>,
}

impl AuthLayer {
    pub fn new(verifier: Arc<Verifier>) -> Self {
        Self { verifier, revocation: None }
    }

    pub fn with_revocation(mut self, r: Arc<RevocationChecker>) -> Self {
        self.revocation = Some(r);
        self
    }
}

impl<S> Layer<S> for AuthLayer {
    type Service = AuthService<S>;
    fn layer(&self, inner: S) -> Self::Service {
        AuthService { inner, layer: self.clone() }
    }
}

#[derive(Clone)]
pub struct AuthService<S> {
    inner: S,
    layer: AuthLayer,
}

impl<S> Service<Request<Body>> for AuthService<S>
where
    S: Service<Request<Body>, Response = Response<Body>> + Clone + Send + 'static,
    S::Future: Send + 'static,
{
    type Response = S::Response;
    type Error = S::Error;
    type Future = std::pin::Pin<Box<dyn std::future::Future<Output = Result<S::Response, S::Error>> + Send>>;

    fn poll_ready(&mut self, cx: &mut Context<'_>) -> Poll<Result<(), Self::Error>> {
        self.inner.poll_ready(cx)
    }

    fn call(&mut self, mut req: Request<Body>) -> Self::Future {
        let layer = self.layer.clone();
        let mut inner = self.inner.clone();
        Box::pin(async move {
            let token = match req.headers().get(http::header::AUTHORIZATION).and_then(|h| h.to_str().ok()) {
                Some(h) if h.starts_with("Bearer ") => h[7..].trim().to_string(),
                _ => return Ok(unauthorized("missing_or_malformed_token")),
            };

            let user = match layer.verifier.verify(&token).await {
                Ok(u) => u,
                Err(_) => return Ok(unauthorized("invalid_token")),
            };

            if let Some(rev) = &layer.revocation {
                if let Some(jti) = user.jti.as_deref() {
                    match rev.is_revoked(jti).await {
                        Ok(true) => return Ok(unauthorized("token_revoked")),
                        _ => {}
                    }
                }
            }

            req.extensions_mut().insert(user);
            inner.call(req).await
        })
    }
}

fn unauthorized(code: &str) -> Response<Body> {
    let body = serde_json::json!({ "error": { "code": code, "message": "unauthorized" } }).to_string();
    let mut resp = (StatusCode::UNAUTHORIZED, body).into_response();
    resp.headers_mut().insert(http::header::CONTENT_TYPE, "application/json".parse().unwrap());
    resp
}
