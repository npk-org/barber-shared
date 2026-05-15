use barber_auth_middleware::Role;

#[test]
fn role_serializes_snake_case() {
    let s = serde_json::to_string(&Role::ShopOwner).unwrap();
    assert_eq!(s, "\"shop_owner\"");

    let r: Role = serde_json::from_str("\"customer\"").unwrap();
    assert_eq!(r, Role::Customer);
}
