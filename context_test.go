package identity

import (
	"context"
	"testing"
)

func TestNew_Valid(t *testing.T) {
	id, err := New("t1", "ws1", "u1", PrincipalUser, "clerk", []string{"admin", "editor"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.TenantID() != "t1" {
		t.Errorf("TenantID = %q, want %q", id.TenantID(), "t1")
	}
	if id.WorkspaceID() != "ws1" {
		t.Errorf("WorkspaceID = %q, want %q", id.WorkspaceID(), "ws1")
	}
	if id.PrincipalID() != "u1" {
		t.Errorf("PrincipalID = %q, want %q", id.PrincipalID(), "u1")
	}
	if id.PrincipalType() != PrincipalUser {
		t.Errorf("PrincipalType = %q, want %q", id.PrincipalType(), PrincipalUser)
	}
	if id.Issuer() != "clerk" {
		t.Errorf("Issuer = %q, want %q", id.Issuer(), "clerk")
	}
	if !id.IsUser() {
		t.Error("IsUser() = false, want true")
	}
	if id.IsService() {
		t.Error("IsService() = true, want false")
	}
}

func TestNew_ValidationErrors(t *testing.T) {
	tests := []struct {
		name      string
		tenant    string
		workspace string
		principal string
		ptype     PrincipalType
		issuer    string
	}{
		{"missing tenant", "", "ws", "u1", PrincipalUser, "clerk"},
		{"missing workspace", "t1", "", "u1", PrincipalUser, "clerk"},
		{"missing principal", "t1", "ws", "", PrincipalUser, "clerk"},
		{"bad principal type", "t1", "ws", "u1", "robot", "clerk"},
		{"missing issuer", "t1", "ws", "u1", PrincipalUser, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.tenant, tt.workspace, tt.principal, tt.ptype, tt.issuer, nil)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestRoles_DefensiveCopy(t *testing.T) {
	original := []string{"admin", "editor"}
	id, _ := New("t1", "ws1", "u1", PrincipalUser, "clerk", original)

	// Mutate the original slice — should not affect the identity.
	original[0] = "hacked"

	// Mutate the returned slice — should not affect the identity.
	roles := id.Roles()
	roles[0] = "hacked"

	if id.Roles()[0] != "admin" {
		t.Error("role slice was mutated through external reference")
	}
}

func TestHasRole(t *testing.T) {
	id, _ := New("t1", "ws1", "u1", PrincipalUser, "clerk", []string{"admin", "editor"})

	if !id.HasRole("admin") {
		t.Error("HasRole(admin) = false, want true")
	}
	if id.HasRole("viewer") {
		t.Error("HasRole(viewer) = true, want false")
	}
}

func TestHasAnyRole(t *testing.T) {
	id, _ := New("t1", "ws1", "u1", PrincipalUser, "clerk", []string{"editor"})

	if !id.HasAnyRole("admin", "editor") {
		t.Error("HasAnyRole(admin, editor) = false, want true")
	}
	if id.HasAnyRole("admin", "viewer") {
		t.Error("HasAnyRole(admin, viewer) = true, want false")
	}
}

func TestContextRoundTrip(t *testing.T) {
	id, _ := New("t1", "ws1", "svc1", PrincipalService, "auth0", []string{"ingester"})

	ctx := WithContext(context.Background(), id)
	got, ok := FromContext(ctx)
	if !ok {
		t.Fatal("FromContext returned ok=false")
	}
	if got.PrincipalID() != "svc1" {
		t.Errorf("PrincipalID = %q, want %q", got.PrincipalID(), "svc1")
	}
}

func TestFromContext_Missing(t *testing.T) {
	_, ok := FromContext(context.Background())
	if ok {
		t.Error("FromContext on empty context returned ok=true")
	}
}

func TestMustFromContext_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustFromContext did not panic on empty context")
		}
	}()
	MustFromContext(context.Background())
}

func TestString(t *testing.T) {
	id, _ := New("t1", "ws1", "u1", PrincipalUser, "clerk", []string{"admin"})
	want := "user:u1@t1/ws1 issuer=clerk roles=[admin]"
	if got := id.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
