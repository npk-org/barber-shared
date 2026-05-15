//! JWT verification middleware for the barber-booking system (Rust services).
//!
//! - `Verifier` fetches JWKS from auth-svc and caches public keys.
//! - `AuthLayer` is a Tower layer that verifies the Bearer token on every
//!   incoming request and inserts a `User` into request extensions.
//! - `User` is an Axum extractor consumers can pull off the request.

pub mod context;
pub mod extractor;
pub mod layer;
pub mod revocation;
pub mod verifier;

pub use context::{Role, User};
pub use layer::{AuthLayer, AuthService};
pub use revocation::RevocationChecker;
pub use verifier::Verifier;
