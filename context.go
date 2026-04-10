// Package identity provides tenant-scoped identity context for multi-tenant services.
//
// Every authenticated request carries an identity [Context] describing who is calling,
// which tenant and workspace they belong to, and what roles they hold. Use the
// context helpers to propagate and retrieve this information through [context.Context].
//
//	ctx = identity.WithContext(ctx, id)
//	id, ok := identity.FromContext(ctx)
package identity

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

// PrincipalType distinguishes between human users and service accounts.
type PrincipalType string

const (
	PrincipalUser    PrincipalType = "user"
	PrincipalService PrincipalType = "service"
)

// Context carries the authenticated caller's identity within a tenant.
//
// All fields are read-only after construction; create instances via [New].
type Context struct {
	tenantID      string
	workspaceID   string
	principalID   string
	principalType PrincipalType
	issuer        string
	roles         []string
}

// New creates a validated identity Context. Returns an error if any required
// field is empty or principalType is not a known value.
func New(tenantID, workspaceID, principalID string, principalType PrincipalType, issuer string, roles []string) (Context, error) {
	if tenantID == "" {
		return Context{}, fmt.Errorf("identity: tenant ID is required")
	}
	if workspaceID == "" {
		return Context{}, fmt.Errorf("identity: workspace ID is required")
	}
	if principalID == "" {
		return Context{}, fmt.Errorf("identity: principal ID is required")
	}
	switch principalType {
	case PrincipalUser, PrincipalService:
	default:
		return Context{}, fmt.Errorf("identity: unknown principal type %q", principalType)
	}
	if issuer == "" {
		return Context{}, fmt.Errorf("identity: issuer is required")
	}

	// Defensive copy so callers can't mutate our slice.
	r := make([]string, len(roles))
	copy(r, roles)

	return Context{
		tenantID:      tenantID,
		workspaceID:   workspaceID,
		principalID:   principalID,
		principalType: principalType,
		issuer:        issuer,
		roles:         r,
	}, nil
}

// Accessors — no setters; identity is immutable once created.

func (c Context) TenantID() string             { return c.tenantID }
func (c Context) WorkspaceID() string          { return c.workspaceID }
func (c Context) PrincipalID() string          { return c.principalID }
func (c Context) PrincipalType() PrincipalType { return c.principalType }
func (c Context) Issuer() string               { return c.issuer }

// Roles returns a copy of the role list.
func (c Context) Roles() []string {
	out := make([]string, len(c.roles))
	copy(out, c.roles)
	return out
}

// HasRole reports whether the identity holds the given role.
func (c Context) HasRole(role string) bool {
	return slices.Contains(c.roles, role)
}

// HasAnyRole reports whether the identity holds at least one of the given roles.
func (c Context) HasAnyRole(roles ...string) bool {
	return slices.ContainsFunc(roles, func(role string) bool {
		return slices.Contains(c.roles, role)
	})
}

// IsUser is a convenience check for PrincipalUser.
func (c Context) IsUser() bool { return c.principalType == PrincipalUser }

// IsService is a convenience check for PrincipalService.
func (c Context) IsService() bool { return c.principalType == PrincipalService }

// String returns a human-readable representation useful for logging.
// Never includes secret material.
func (c Context) String() string {
	return fmt.Sprintf("%s:%s@%s/%s issuer=%s roles=[%s]",
		c.principalType, c.principalID,
		c.tenantID, c.workspaceID,
		c.issuer, strings.Join(c.roles, ","),
	)
}

// LogValue implements slog.LogValuer so the identity renders cleanly
// in structured log output.
func (c Context) LogValue() map[string]any {
	return map[string]any{
		"tenant_id":      c.tenantID,
		"workspace_id":   c.workspaceID,
		"principal_id":   c.principalID,
		"principal_type": string(c.principalType),
		"issuer":         c.issuer,
		"roles":          c.roles,
	}
}

// --- context.Context integration ---

type contextKey struct{}

// WithContext returns a new context carrying the given identity.
func WithContext(ctx context.Context, id Context) context.Context {
	return context.WithValue(ctx, contextKey{}, id)
}

// FromContext extracts the identity from ctx. Returns the zero Context and
// false if none is present.
func FromContext(ctx context.Context) (Context, bool) {
	id, ok := ctx.Value(contextKey{}).(Context)
	return id, ok
}

// MustFromContext extracts the identity from ctx, panicking if absent.
// Use only in code paths where middleware guarantees the identity exists.
func MustFromContext(ctx context.Context) Context {
	id, ok := FromContext(ctx)
	if !ok {
		panic("identity: no identity in context — is the auth middleware configured?")
	}
	return id
}
