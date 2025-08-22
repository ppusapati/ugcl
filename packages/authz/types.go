package authz

import pb "p9e.in/ugcl/identity/api/v2/permission"

// InjectedUserInfo holds parsed user claims injected into the gRPC context
// after token verification.
type InjectedUserInfo struct {
	UserID      string
	TenantID    string
	Role        string
	Permissions []pb.Permission // Optional: if JWT bundling is enabled
}

// PermissionRequirement defines the required namespace/resource/action
// for a given RPC method or protected resource.
type PermissionRequirement struct {
	Namespace string
	Resource  string
	Action    string
}
