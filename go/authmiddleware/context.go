package authmiddleware

import "context"

type Role string

const (
	RoleCustomer  Role = "customer"
	RoleBarber    Role = "barber"
	RoleShopOwner Role = "shop_owner"
	RoleAdmin     Role = "admin"
)

type User struct {
	ID     string
	Role   Role
	ShopID string
	JTI    string
}

type ctxKey struct{}

func WithUser(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, ctxKey{}, u)
}

func FromContext(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(ctxKey{}).(User)
	return u, ok
}
