package authmiddleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExtractBearer(t *testing.T) {
	cases := []struct {
		name   string
		header string
		want   string
		err    bool
	}{
		{"empty", "", "", true},
		{"no_prefix", "abc.def.ghi", "", true},
		{"valid", "Bearer abc.def.ghi", "abc.def.ghi", false},
		{"valid_padded", "Bearer   abc.def.ghi  ", "abc.def.ghi", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r, _ := http.NewRequest("GET", "/", nil)
			if c.header != "" {
				r.Header.Set("Authorization", c.header)
			}
			got, err := extractBearer(r)
			if (err != nil) != c.err {
				t.Fatalf("err=%v wantErr=%v", err, c.err)
			}
			if got != c.want {
				t.Fatalf("got=%q want=%q", got, c.want)
			}
		})
	}
}

func TestUserContextRoundtrip(t *testing.T) {
	u := User{ID: "usr_1", Role: RoleBarber, ShopID: "shp_1", JTI: "j1"}
	ctx := WithUser(context.Background(), u)
	got, ok := FromContext(ctx)
	if !ok {
		t.Fatal("expected user in context")
	}
	if got != u {
		t.Fatalf("got=%+v want=%+v", got, u)
	}
}

func TestRequireRole(t *testing.T) {
	handler := RequireRole(RoleBarber, RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	cases := []struct {
		name     string
		user     *User
		wantCode int
	}{
		{"no_user", nil, http.StatusUnauthorized},
		{"wrong_role", &User{ID: "u", Role: RoleCustomer}, http.StatusForbidden},
		{"allowed_barber", &User{ID: "u", Role: RoleBarber}, http.StatusOK},
		{"allowed_admin", &User{ID: "u", Role: RoleAdmin}, http.StatusOK},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r, _ := http.NewRequest("GET", "/", nil)
			if c.user != nil {
				r = r.WithContext(WithUser(r.Context(), *c.user))
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			if w.Code != c.wantCode {
				t.Fatalf("got=%d want=%d", w.Code, c.wantCode)
			}
		})
	}
}
