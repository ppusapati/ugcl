# Proto-First Development Guide

## Overview

UGCL follows a **proto-first development** approach where API contracts are defined in `.proto` files before implementation.

### Workflow

```
1. Define service in .proto file
   ↓
2. Run buf generate (creates Go stubs)
   ↓
3. Implement service interface
   ↓
4. Register handler in module.go
   ↓
5. Test with Connect client
```

### Benefits

- **Contract-First:** API defined before implementation
- **Type Safety:** Strong typing across services
- **Auto-Generation:** Less boilerplate code
- **Cross-Language:** Works with multiple languages

## Quick Start

**1. Define proto:**
```protobuf
// proto/user.proto
syntax = "proto3";

package user.v1;

service UserService {
  rpc CreateUser(CreateUserRequest) returns (User);
}

message User {
  string id = 1;
  string email = 2;
}
```

**2. Generate code:**
```bash
buf generate
```

**3. Implement:**
```go
func (h *UserHandler) CreateUser(
    ctx context.Context,
    req *connect.Request[userv1.CreateUserRequest],
) (*connect.Response[userv1.User], error) {
    // Implementation
}
```

**4. Register:**
```go
fx.Provide(
    handlers.NewUserHandler,
    func(h *handlers.UserHandler) (string, http.Handler) {
        return userv1connect.NewUserServiceHandler(h)
    },
)
```

## Proto File Structure

```protobuf
syntax = "proto3";

package {module}.v1;

option go_package = "p9e.in/ugcl/{module}/api/v1/{module}";

// Service definition
service {Module}Service {
  rpc Create{Resource}(Create{Resource}Request) returns ({Resource}) {}
  rpc Get{Resource}(Get{Resource}Request) returns ({Resource}) {}
  rpc List{Resources}(List{Resources}Request) returns (List{Resources}Response) {}
  rpc Update{Resource}(Update{Resource}Request) returns ({Resource}) {}
  rpc Delete{Resource}(Delete{Resource}Request) returns (Delete{Resource}Response) {}
}
```

## Best Practices

### Naming
- Services: `PascalCase` (e.g., `UserService`)
- RPCs: `PascalCase` (e.g., `GetUser`)
- Messages: `PascalCase` (e.g., `User`)
- Fields: `snake_case` (e.g., `user_id`)

### Versioning
```protobuf
package user.v1;  // Include version
option go_package = "p9e.in/ugcl/user/api/v1/user";
```

### Error Handling
```go
// Use Connect error codes
return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found"))
return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid email"))
return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("access denied"))
```

### Pagination
```protobuf
message ListUsersRequest {
  int32 page = 1;
  int32 page_size = 2;
  optional string filter = 3;
}

message ListUsersResponse {
  repeated User users = 1;
  int32 total_count = 2;
}
```

## Code Generation

**Buf configuration (buf.gen.yaml):**
```yaml
version: v1
plugins:
  - plugin: buf.build/protocolbuffers/go
    out: .
    opt:
      - paths=source_relative
  - plugin: buf.build/connectrpc/go
    out: .
    opt:
      - paths=source_relative
```

**Generate:**
```bash
# Generate all
buf generate

# Generate specific file
buf generate --path proto/user.proto

# Check breaking changes
buf breaking --against '.git#branch=main'
```

## Additional Resources

- [Protocol Buffers](https://protobuf.dev/)
- [Buf Documentation](https://docs.buf.build/)
- [Connect-RPC](https://connectrpc.com/docs/go/)
- [Proto Style Guide](https://protobuf.dev/programming-guides/style/)
