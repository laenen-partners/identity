# identity

Tenant-scoped identity context for Go multi-tenant services.

`identity` provides an immutable, validated `Context` type that travels through `context.Context`, giving every layer of your application access to **who** is calling, **which tenant** they belong to, and **what they can do**.

## Install

```sh
go get github.com/laenen-partners/identity
```

## Quick start

### Create an identity

```go
id, err := identity.New(
    "tenant_abc",           // tenant ID
    "ws_prod",              // workspace ID
    "usr_42",               // principal ID
    identity.PrincipalUser, // "user" or "service"
    []string{"admin", "editor"},
)
if err != nil {
    // handles missing fields or unknown principal type
}
```

### Propagate through context

```go
// Attach to context (typically in auth middleware)
ctx = identity.WithContext(ctx, id)

// Retrieve downstream
id, ok := identity.FromContext(ctx)

// Or panic if absent — use only behind auth middleware
id := identity.MustFromContext(ctx)
```

### Check roles

```go
if id.HasRole("admin") {
    // ...
}

if id.HasAnyRole("admin", "editor") {
    // ...
}
```

### Structured logging

`Context` implements `String()` and `LogValue()` for clean output with `slog`:

```go
slog.Info("request authorized", "identity", id)
// => identity.tenant_id=tenant_abc identity.workspace_id=ws_prod ...
```

## API overview

| Method | Description |
|---|---|
| `New(...)` | Validated constructor; returns error on invalid input |
| `TenantID()` | Tenant the caller belongs to |
| `WorkspaceID()` | Workspace within the tenant |
| `PrincipalID()` | User or service account identifier |
| `PrincipalType()` | `PrincipalUser` or `PrincipalService` |
| `Roles()` | Copy of the caller's role list |
| `HasRole(role)` | Check for a single role |
| `HasAnyRole(roles...)` | Check for at least one of the given roles |
| `IsUser()` / `IsService()` | Convenience type checks |
| `WithContext(ctx, id)` | Store identity in `context.Context` |
| `FromContext(ctx)` | Retrieve identity (returns `ok` bool) |
| `MustFromContext(ctx)` | Retrieve identity or panic |

## Design decisions

- **Immutable after construction.** Fields are unexported with read-only accessors. Role slices are defensively copied on input and output. This prevents accidental or malicious mutation of security-critical data.
- **Validated at the boundary.** `New` rejects empty IDs and unknown principal types so downstream code can trust the values without re-checking.
- **Zero dependencies.** Standard library only.

## Testing

```sh
go test -v -count=1 ./...
```

## License

MIT
