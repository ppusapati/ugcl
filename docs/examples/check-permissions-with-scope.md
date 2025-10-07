# Checking Permissions with Scope - Complete Guide

This comprehensive guide demonstrates permission checking workflows in the UGCL Backend v2 system, including entity-based permissions, organizational scope matching, hierarchical inheritance, resource-level permissions, and the complete 5-step resolution process.

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Client Setup](#client-setup)
4. [Basic Permission Checks](#basic-permission-checks)
5. [Entity-Based Permission Checks](#entity-based-permission-checks)
6. [Organizational Scope Matching](#organizational-scope-matching)
7. [Hierarchical Permission Inheritance](#hierarchical-permission-inheritance)
8. [Resource-Level Permissions](#resource-level-permissions)
9. [Permission Resolution Path](#permission-resolution-path)
10. [Temporal Permission Validation](#temporal-permission-validation)
11. [Multiple Permission Scenarios](#multiple-permission-scenarios)
12. [Complete 5-Step Resolution](#complete-5-step-resolution)
13. [Integration with Auth Middleware](#integration-with-auth-middleware)
14. [Error Handling](#error-handling)
15. [Best Practices](#best-practices)

## Overview

The UGCL permission system uses a 5-step resolution process:

1. **Direct Permissions**: Explicitly granted to the subject
2. **Entity-Based Permissions**: Via entity role bindings
3. **Role-Based Permissions**: Via assigned roles
4. **Inherited Permissions**: From parent organizational units
5. **Resource Ownership**: Based on ownership or sharing

## Prerequisites

```go
import (
    "context"
    "fmt"
    "log"
    "time"

    "connectrpc.com/connect"
    permissionv1 "p9e.in/ugcl/identity/user/api/v2/permission"
    "p9e.in/ugcl/identity/user/api/v2/permission/permissionconnect"
    "google.golang.org/protobuf/types/known/timestamppb"
    "google.golang.org/protobuf/types/known/wrapperspb"
)
```

## Client Setup

### Permission Client Configuration

```go
package main

import (
    "crypto/tls"
    "net/http"
    "time"
)

// PermissionClient encapsulates the permission service client
type PermissionClient struct {
    client   permissionconnect.PermissionServiceClient
    tenantID string
    userID   string
}

// NewPermissionClient creates a new permission service client
func NewPermissionClient(baseURL, tenantID, userID, authToken string) *PermissionClient {
    // Configure HTTP client
    httpClient := &http.Client{
        Timeout: 10 * time.Second,
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                MinVersion: tls.VersionTLS12,
            },
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 100,
            IdleConnTimeout:     90 * time.Second,
        },
    }

    // Create Connect client
    client := permissionconnect.NewPermissionServiceClient(
        httpClient,
        baseURL,
        connect.WithInterceptors(
            authInterceptor(authToken),
            loggingInterceptor(),
        ),
    )

    return &PermissionClient{
        client:   client,
        tenantID: tenantID,
        userID:   userID,
    }
}

// authInterceptor adds authentication to requests
func authInterceptor(token string) connect.UnaryInterceptorFunc {
    return func(next connect.UnaryFunc) connect.UnaryFunc {
        return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
            req.Header().Set("Authorization", "Bearer "+token)
            return next(ctx, req)
        }
    }
}

// loggingInterceptor logs permission checks
func loggingInterceptor() connect.UnaryInterceptorFunc {
    return func(next connect.UnaryFunc) connect.UnaryFunc {
        return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
            start := time.Now()
            resp, err := next(ctx, req)
            duration := time.Since(start)

            if err == nil {
                log.Printf("Permission check completed in %v", duration)
            }
            return resp, err
        }
    }
}
```

## Basic Permission Checks

### Simple Permission Check

```go
// CheckPermission performs a basic permission check
func (c *PermissionClient) CheckPermission(
    ctx context.Context,
    subject, namespace, resource, action string,
) (bool, error) {
    req := &permissionv1.CheckPermissionRequest{
        Subject:   subject,
        Namespace: namespace,
        Resource:  resource,
        Action:    action,
        TenantId:  c.tenantID,
        Mode:      permissionv1.PermissionCheckMode_PERMISSION_CHECK_MODE_WITH_INHERITANCE,
    }

    resp, err := c.client.Check(ctx, connect.NewRequest(req))
    if err != nil {
        return false, fmt.Errorf("permission check failed: %w", err)
    }

    granted := resp.Msg.Effect == permissionv1.Effect_GRANT
    log.Printf("Permission check: %s.%s.%s for %s = %v (reason: %s)",
        namespace, resource, action, subject, granted, resp.Msg.Reason)

    return granted, nil
}
```

### Permission Check with Result Details

```go
// CheckPermissionDetailed performs a permission check and returns full details
func (c *PermissionClient) CheckPermissionDetailed(
    ctx context.Context,
    subject, namespace, resource, action string,
) (*permissionv1.CheckPermissionResponse, error) {
    req := &permissionv1.CheckPermissionRequest{
        Subject:   subject,
        Namespace: namespace,
        Resource:  resource,
        Action:    action,
        TenantId:  c.tenantID,
        Mode:      permissionv1.PermissionCheckMode_PERMISSION_CHECK_MODE_WITH_INHERITANCE,
    }

    resp, err := c.client.Check(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("permission check failed: %w", err)
    }

    return resp.Msg, nil
}
```

## Entity-Based Permission Checks

### Check Entity Permission

```go
// CheckEntityPermission checks permission for an entity
func (c *PermissionClient) CheckEntityPermission(
    ctx context.Context,
    entityID, namespace, resource, action string,
) (bool, *permissionv1.CheckPermissionResponse, error) {
    subject := fmt.Sprintf("entity:%s", entityID)

    req := &permissionv1.CheckPermissionRequest{
        Subject:   subject,
        Namespace: namespace,
        Resource:  resource,
        Action:    action,
        TenantId:  c.tenantID,
        Mode:      permissionv1.PermissionCheckMode_PERMISSION_CHECK_MODE_WITH_INHERITANCE,
    }

    resp, err := c.client.Check(ctx, connect.NewRequest(req))
    if err != nil {
        return false, nil, fmt.Errorf("entity permission check failed: %w", err)
    }

    granted := resp.Msg.Effect == permissionv1.Effect_GRANT
    return granted, resp.Msg, nil
}
```

### Check Multiple Permissions for Entity

```go
// PermissionCheck represents a permission to check
type PermissionCheck struct {
    Namespace string
    Resource  string
    Action    string
}

// CheckMultiplePermissions checks multiple permissions at once
func (c *PermissionClient) CheckMultiplePermissions(
    ctx context.Context,
    entityID string,
    permissions []PermissionCheck,
) (map[string]bool, error) {
    results := make(map[string]bool)

    for _, perm := range permissions {
        key := fmt.Sprintf("%s.%s.%s", perm.Namespace, perm.Resource, perm.Action)

        granted, _, err := c.CheckEntityPermission(
            ctx,
            entityID,
            perm.Namespace,
            perm.Resource,
            perm.Action,
        )

        if err != nil {
            log.Printf("Warning: Failed to check %s: %v", key, err)
            results[key] = false
            continue
        }

        results[key] = granted
    }

    return results, nil
}
```

## Organizational Scope Matching

### Check Permission with Division Scope

```go
// CheckPermissionWithDivision checks permission scoped to a division
func (c *PermissionClient) CheckPermissionWithDivision(
    ctx context.Context,
    entityID, namespace, resource, action, divisionID string,
) (bool, error) {
    req := &permissionv1.CheckPermissionRequest{
        Subject:    fmt.Sprintf("entity:%s", entityID),
        Namespace:  namespace,
        Resource:   resource,
        Action:     action,
        TenantId:   c.tenantID,
        DivisionId: wrapperspb.String(divisionID),
        Mode:       permissionv1.PermissionCheckMode_PERMISSION_CHECK_MODE_WITH_INHERITANCE,
    }

    resp, err := c.client.Check(ctx, connect.NewRequest(req))
    if err != nil {
        return false, fmt.Errorf("division-scoped permission check failed: %w", err)
    }

    granted := resp.Msg.Effect == permissionv1.Effect_GRANT
    log.Printf("Division-scoped permission: %s in division %s = %v",
        action, divisionID, granted)

    return granted, nil
}
```

### Check Permission with Branch Scope

```go
// CheckPermissionWithBranch checks permission scoped to a branch
func (c *PermissionClient) CheckPermissionWithBranch(
    ctx context.Context,
    entityID, namespace, resource, action, branchID string,
) (bool, error) {
    req := &permissionv1.CheckPermissionRequest{
        Subject:   fmt.Sprintf("entity:%s", entityID),
        Namespace: namespace,
        Resource:  resource,
        Action:    action,
        TenantId:  c.tenantID,
        BranchId:  wrapperspb.String(branchID),
        Mode:      permissionv1.PermissionCheckMode_PERMISSION_CHECK_MODE_WITH_INHERITANCE,
    }

    resp, err := c.client.Check(ctx, connect.NewRequest(req))
    if err != nil {
        return false, fmt.Errorf("branch-scoped permission check failed: %w", err)
    }

    granted := resp.Msg.Effect == permissionv1.Effect_GRANT
    return granted, nil
}
```

### Check Permission with Department Scope

```go
// CheckPermissionWithDepartment checks permission scoped to a department
func (c *PermissionClient) CheckPermissionWithDepartment(
    ctx context.Context,
    entityID, namespace, resource, action, departmentID string,
) (bool, error) {
    req := &permissionv1.CheckPermissionRequest{
        Subject:      fmt.Sprintf("entity:%s", entityID),
        Namespace:    namespace,
        Resource:     resource,
        Action:       action,
        TenantId:     c.tenantID,
        DepartmentId: wrapperspb.String(departmentID),
        Mode:         permissionv1.PermissionCheckMode_PERMISSION_CHECK_MODE_WITH_INHERITANCE,
    }

    resp, err := c.client.Check(ctx, connect.NewRequest(req))
    if err != nil {
        return false, fmt.Errorf("department-scoped permission check failed: %w", err)
    }

    granted := resp.Msg.Effect == permissionv1.Effect_GRANT
    return granted, nil
}
```

### Check Permission with Full Organizational Context

```go
// CheckPermissionWithOrgContext checks permission with complete org hierarchy
func (c *PermissionClient) CheckPermissionWithOrgContext(
    ctx context.Context,
    entityID, namespace, resource, action string,
    divisionID, branchID, departmentID string,
) (*permissionv1.CheckPermissionResponse, error) {
    req := &permissionv1.CheckPermissionRequest{
        Subject:   fmt.Sprintf("entity:%s", entityID),
        Namespace: namespace,
        Resource:  resource,
        Action:    action,
        TenantId:  c.tenantID,
        Mode:      permissionv1.PermissionCheckMode_PERMISSION_CHECK_MODE_WITH_INHERITANCE,
    }

    if divisionID != "" {
        req.DivisionId = wrapperspb.String(divisionID)
    }
    if branchID != "" {
        req.BranchId = wrapperspb.String(branchID)
    }
    if departmentID != "" {
        req.DepartmentId = wrapperspb.String(departmentID)
    }

    resp, err := c.client.Check(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("org-scoped permission check failed: %w", err)
    }

    return resp.Msg, nil
}
```

## Hierarchical Permission Inheritance

### Check Permission with Inheritance

```go
// CheckWithInheritance checks permission allowing inheritance from parent units
func (c *PermissionClient) CheckWithInheritance(
    ctx context.Context,
    entityID, namespace, resource, action string,
    departmentID string,
) (*permissionv1.CheckPermissionResponse, error) {
    req := &permissionv1.CheckPermissionRequest{
        Subject:      fmt.Sprintf("entity:%s", entityID),
        Namespace:    namespace,
        Resource:     resource,
        Action:       action,
        TenantId:     c.tenantID,
        DepartmentId: wrapperspb.String(departmentID),
        Mode:         permissionv1.PermissionCheckMode_PERMISSION_CHECK_MODE_WITH_INHERITANCE,
    }

    resp, err := c.client.Check(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("inheritance check failed: %w", err)
    }

    // Log inheritance path
    if len(resp.Msg.ResolutionPath) > 0 {
        log.Println("Permission resolution path:")
        for i, step := range resp.Msg.ResolutionPath {
            log.Printf("  %d. %s (source: %s, scope: %s)",
                i+1, step.Description, step.Source, step.Scope)
        }
    }

    return resp.Msg, nil
}
```

### Check Exact Match Only (No Inheritance)

```go
// CheckExactMatch checks for exact permission match without inheritance
func (c *PermissionClient) CheckExactMatch(
    ctx context.Context,
    entityID, namespace, resource, action string,
    divisionID, branchID, departmentID string,
) (bool, error) {
    req := &permissionv1.CheckPermissionRequest{
        Subject:   fmt.Sprintf("entity:%s", entityID),
        Namespace: namespace,
        Resource:  resource,
        Action:    action,
        TenantId:  c.tenantID,
        Mode:      permissionv1.PermissionCheckMode_PERMISSION_CHECK_MODE_EXACT,
    }

    if divisionID != "" {
        req.DivisionId = wrapperspb.String(divisionID)
    }
    if branchID != "" {
        req.BranchId = wrapperspb.String(branchID)
    }
    if departmentID != "" {
        req.DepartmentId = wrapperspb.String(departmentID)
    }

    resp, err := c.client.Check(ctx, connect.NewRequest(req))
    if err != nil {
        return false, fmt.Errorf("exact match check failed: %w", err)
    }

    granted := resp.Msg.Effect == permissionv1.Effect_GRANT
    return granted, nil
}
```

## Resource-Level Permissions

### Check Permission on Specific Resource Instance

```go
// CheckResourcePermission checks permission on a specific resource instance
func (c *PermissionClient) CheckResourcePermission(
    ctx context.Context,
    entityID, namespace, resource, action, resourceID string,
) (bool, *permissionv1.CheckPermissionResponse, error) {
    req := &permissionv1.CheckPermissionRequest{
        Subject:    fmt.Sprintf("entity:%s", entityID),
        Namespace:  namespace,
        Resource:   resource,
        Action:     action,
        TenantId:   c.tenantID,
        ResourceId: wrapperspb.String(resourceID),
        Mode:       permissionv1.PermissionCheckMode_PERMISSION_CHECK_MODE_WITH_INHERITANCE,
    }

    resp, err := c.client.Check(ctx, connect.NewRequest(req))
    if err != nil {
        return false, nil, fmt.Errorf("resource permission check failed: %w", err)
    }

    granted := resp.Msg.Effect == permissionv1.Effect_GRANT
    log.Printf("Resource permission: %s on %s/%s = %v",
        action, resource, resourceID, granted)

    return granted, resp.Msg, nil
}
```

### Check Document Access Permission

```go
// CheckDocumentAccess checks if entity can access a specific document
func (c *PermissionClient) CheckDocumentAccess(
    ctx context.Context,
    entityID, documentID, action string,
) (bool, string, error) {
    req := &permissionv1.CheckPermissionRequest{
        Subject:    fmt.Sprintf("entity:%s", entityID),
        Namespace:  "dms",
        Resource:   "document",
        Action:     action,
        TenantId:   c.tenantID,
        ResourceId: wrapperspb.String(documentID),
        Mode:       permissionv1.PermissionCheckMode_PERMISSION_CHECK_MODE_WITH_INHERITANCE,
    }

    resp, err := c.client.Check(ctx, connect.NewRequest(req))
    if err != nil {
        return false, "", fmt.Errorf("document access check failed: %w", err)
    }

    granted := resp.Msg.Effect == permissionv1.Effect_GRANT
    return granted, resp.Msg.Reason, nil
}
```

### Check Multiple Resource Actions

```go
// CheckDocumentPermissions checks all common document permissions
func (c *PermissionClient) CheckDocumentPermissions(
    ctx context.Context,
    entityID, documentID string,
) (map[string]bool, error) {
    actions := []string{"view", "download", "edit", "delete", "share"}
    results := make(map[string]bool)

    for _, action := range actions {
        granted, _, err := c.CheckDocumentAccess(ctx, entityID, documentID, action)
        if err != nil {
            log.Printf("Warning: Failed to check %s permission: %v", action, err)
            results[action] = false
            continue
        }
        results[action] = granted
    }

    log.Printf("Document permissions for %s:", documentID)
    for action, granted := range results {
        log.Printf("  %s: %v", action, granted)
    }

    return results, nil
}
```

## Permission Resolution Path

### Analyze Resolution Path

```go
// AnalyzeResolutionPath analyzes how a permission was resolved
func (c *PermissionClient) AnalyzeResolutionPath(
    ctx context.Context,
    entityID, namespace, resource, action string,
) error {
    req := &permissionv1.CheckPermissionRequest{
        Subject:   fmt.Sprintf("entity:%s", entityID),
        Namespace: namespace,
        Resource:  resource,
        Action:    action,
        TenantId:  c.tenantID,
        Mode:      permissionv1.PermissionCheckMode_PERMISSION_CHECK_MODE_WITH_INHERITANCE,
    }

    resp, err := c.client.Check(ctx, connect.NewRequest(req))
    if err != nil {
        return fmt.Errorf("permission check failed: %w", err)
    }

    log.Printf("=== Permission Resolution Analysis ===")
    log.Printf("Permission: %s.%s.%s", namespace, resource, action)
    log.Printf("Subject: entity:%s", entityID)
    log.Printf("Result: %s", resp.Effect)
    log.Printf("Reason: %s", resp.Reason)

    if len(resp.ResolutionPath) > 0 {
        log.Printf("\nResolution Path (%d steps):", len(resp.ResolutionPath))
        for i, step := range resp.ResolutionPath {
            log.Printf("  Step %d:", i+1)
            log.Printf("    Source: %s", step.Source)
            log.Printf("    Source ID: %s", step.SourceId)
            log.Printf("    Scope: %s", step.Scope)
            log.Printf("    Description: %s", step.Description)
        }
    } else {
        log.Println("\nNo resolution path (permission denied)")
    }

    if resp.MatchedPermission != nil {
        log.Printf("\nMatched Permission:")
        log.Printf("  Namespace: %s", resp.MatchedPermission.Namespace)
        log.Printf("  Resource: %s", resp.MatchedPermission.Resource)
        log.Printf("  Action: %s", resp.MatchedPermission.Action)
        log.Printf("  Effect: %s", resp.MatchedPermission.Effect)
        if resp.MatchedPermission.DivisionId != nil {
            log.Printf("  Division: %s", resp.MatchedPermission.DivisionId.Value)
        }
        if resp.MatchedPermission.BranchId != nil {
            log.Printf("  Branch: %s", resp.MatchedPermission.BranchId.Value)
        }
        if resp.MatchedPermission.DepartmentId != nil {
            log.Printf("  Department: %s", resp.MatchedPermission.DepartmentId.Value)
        }
    }

    return nil
}
```

## Temporal Permission Validation

### Check Time-Bound Permission

```go
// CheckTemporalPermission checks if a permission is valid at a specific time
func (c *PermissionClient) CheckTemporalPermission(
    ctx context.Context,
    entityID, namespace, resource, action string,
    checkTime time.Time,
) (bool, error) {
    // Note: The actual implementation would need to pass checkTime to the service
    // This is a simplified version showing the concept

    req := &permissionv1.CheckPermissionRequest{
        Subject:   fmt.Sprintf("entity:%s", entityID),
        Namespace: namespace,
        Resource:  resource,
        Action:    action,
        TenantId:  c.tenantID,
        Mode:      permissionv1.PermissionCheckMode_PERMISSION_CHECK_MODE_WITH_INHERITANCE,
    }

    resp, err := c.client.Check(ctx, connect.NewRequest(req))
    if err != nil {
        return false, fmt.Errorf("temporal permission check failed: %w", err)
    }

    granted := resp.Msg.Effect == permissionv1.Effect_GRANT

    // Check if matched permission has temporal constraints
    if granted && resp.Msg.MatchedPermission != nil {
        perm := resp.Msg.MatchedPermission

        if perm.ValidFrom != nil {
            validFrom := perm.ValidFrom.AsTime()
            if checkTime.Before(validFrom) {
                log.Printf("Permission not yet valid (valid from: %s)", validFrom)
                return false, nil
            }
        }

        if perm.ValidUntil != nil {
            validUntil := perm.ValidUntil.AsTime()
            if checkTime.After(validUntil) {
                log.Printf("Permission expired (valid until: %s)", validUntil)
                return false, nil
            }
        }
    }

    return granted, nil
}
```

### Check Permission Expiration

```go
// CheckPermissionExpiration checks when a permission will expire
func (c *PermissionClient) CheckPermissionExpiration(
    ctx context.Context,
    entityID, namespace, resource, action string,
) (*time.Time, error) {
    req := &permissionv1.CheckPermissionRequest{
        Subject:   fmt.Sprintf("entity:%s", entityID),
        Namespace: namespace,
        Resource:  resource,
        Action:    action,
        TenantId:  c.tenantID,
        Mode:      permissionv1.PermissionCheckMode_PERMISSION_CHECK_MODE_WITH_INHERITANCE,
    }

    resp, err := c.client.Check(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("permission check failed: %w", err)
    }

    if resp.Msg.Effect != permissionv1.Effect_GRANT {
        return nil, fmt.Errorf("permission not granted")
    }

    if resp.Msg.MatchedPermission != nil && resp.Msg.MatchedPermission.ValidUntil != nil {
        expiration := resp.Msg.MatchedPermission.ValidUntil.AsTime()
        return &expiration, nil
    }

    return nil, nil // No expiration
}
```

## Multiple Permission Scenarios

### Scenario: Document Management

```go
// CheckDocumentManagementPermissions checks all document-related permissions
func (c *PermissionClient) CheckDocumentManagementPermissions(
    ctx context.Context,
    entityID, divisionID, departmentID string,
) error {
    log.Println("=== Document Management Permissions ===")

    // Scenario 1: Create document in department
    canCreate, err := c.CheckPermissionWithDepartment(
        ctx,
        entityID,
        "dms",
        "document",
        "create",
        departmentID,
    )
    if err != nil {
        return err
    }
    log.Printf("Create document in department: %v", canCreate)

    // Scenario 2: View all documents in division
    canViewAll, err := c.CheckPermissionWithDivision(
        ctx,
        entityID,
        "dms",
        "document",
        "view",
        divisionID,
    )
    if err != nil {
        return err
    }
    log.Printf("View all documents in division: %v", canViewAll)

    // Scenario 3: Manage watermarks
    canManageWatermarks, err := c.CheckPermission(
        ctx,
        fmt.Sprintf("entity:%s", entityID),
        "dms",
        "watermark",
        "manage",
    )
    if err != nil {
        return err
    }
    log.Printf("Manage watermarks: %v", canManageWatermarks)

    return nil
}
```

### Scenario: HR Operations

```go
// CheckHRPermissions checks HR-related permissions
func (c *PermissionClient) CheckHRPermissions(
    ctx context.Context,
    entityID, divisionID string,
) (map[string]bool, error) {
    permissions := map[string]bool{}

    // Check various HR operations
    hrChecks := []struct {
        resource string
        action   string
        label    string
    }{
        {"employee", "create", "Create Employee"},
        {"employee", "update", "Update Employee"},
        {"employee", "view", "View Employee"},
        {"employee", "delete", "Delete Employee"},
        {"payroll", "process", "Process Payroll"},
        {"attendance", "manage", "Manage Attendance"},
        {"leave", "approve", "Approve Leave"},
    }

    for _, check := range hrChecks {
        granted, err := c.CheckPermissionWithDivision(
            ctx,
            entityID,
            "hr",
            check.resource,
            check.action,
            divisionID,
        )
        if err != nil {
            log.Printf("Warning: Failed to check %s: %v", check.label, err)
            permissions[check.label] = false
            continue
        }
        permissions[check.label] = granted
    }

    return permissions, nil
}
```

## Complete 5-Step Resolution

### Demonstrate All Resolution Steps

```go
// DemonstrateResolutionSteps shows all 5 permission resolution steps
func (c *PermissionClient) DemonstrateResolutionSteps(
    ctx context.Context,
    entityID string,
) error {
    log.Println("=== Permission Resolution Steps Demo ===")

    // Step 1: Direct Permission
    log.Println("\n1. Checking direct permissions...")
    resp1, err := c.CheckPermissionDetailed(
        ctx,
        fmt.Sprintf("entity:%s", entityID),
        "test",
        "resource",
        "direct-action",
    )
    if err != nil {
        log.Printf("Error: %v", err)
    } else {
        log.Printf("Result: %s - %s", resp1.Effect, resp1.Reason)
    }

    // Step 2: Entity Role-Based Permission
    log.Println("\n2. Checking entity role permissions...")
    resp2, err := c.CheckPermissionDetailed(
        ctx,
        fmt.Sprintf("entity:%s", entityID),
        "dms",
        "document",
        "create",
    )
    if err != nil {
        log.Printf("Error: %v", err)
    } else {
        log.Printf("Result: %s - %s", resp2.Effect, resp2.Reason)
        if len(resp2.ResolutionPath) > 0 {
            for _, step := range resp2.ResolutionPath {
                if step.Source == "entity_role" {
                    log.Printf("  Granted via entity role: %s", step.Description)
                }
            }
        }
    }

    // Step 3: Standard Role Permission
    log.Println("\n3. Checking role permissions...")
    resp3, err := c.CheckPermissionDetailed(
        ctx,
        fmt.Sprintf("entity:%s", entityID),
        "organization",
        "division",
        "view",
    )
    if err != nil {
        log.Printf("Error: %v", err)
    } else {
        log.Printf("Result: %s - %s", resp3.Effect, resp3.Reason)
    }

    // Step 4: Inherited Permission
    log.Println("\n4. Checking inherited permissions...")
    resp4, err := c.CheckWithInheritance(
        ctx,
        entityID,
        "dms",
        "document",
        "view",
        "dept-engineering-001",
    )
    if err != nil {
        log.Printf("Error: %v", err)
    } else {
        log.Printf("Result: %s - %s", resp4.Effect, resp4.Reason)
        if len(resp4.ResolutionPath) > 0 {
            for _, step := range resp4.ResolutionPath {
                if step.Source == "inherited" {
                    log.Printf("  Inherited from %s level", step.Scope)
                }
            }
        }
    }

    // Step 5: Resource Ownership
    log.Println("\n5. Checking resource ownership...")
    documentID := "doc-owned-by-entity"
    resp5, _, err := c.CheckResourcePermission(
        ctx,
        entityID,
        "dms",
        "document",
        "edit",
        documentID,
    )
    if err != nil {
        log.Printf("Error: %v", err)
    } else {
        log.Printf("Result: %v", resp5)
    }

    return nil
}
```

## Integration with Auth Middleware

### Auth Middleware Implementation

```go
// AuthMiddleware provides permission checking for HTTP handlers
type AuthMiddleware struct {
    permissionClient *PermissionClient
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(client *PermissionClient) *AuthMiddleware {
    return &AuthMiddleware{
        permissionClient: client,
    }
}

// RequirePermission creates a middleware that checks for a specific permission
func (m *AuthMiddleware) RequirePermission(namespace, resource, action string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract entity ID from request context (set by auth handler)
            entityID := r.Context().Value("entity_id").(string)

            // Check permission
            granted, err := m.permissionClient.CheckPermission(
                r.Context(),
                fmt.Sprintf("entity:%s", entityID),
                namespace,
                resource,
                action,
            )

            if err != nil {
                http.Error(w, "Permission check failed", http.StatusInternalServerError)
                return
            }

            if !granted {
                http.Error(w, "Permission denied", http.StatusForbidden)
                return
            }

            // Permission granted, continue
            next.ServeHTTP(w, r)
        })
    }
}
```

### Route Protection Example

```go
// ProtectRoutes demonstrates protecting HTTP routes with permissions
func ProtectRoutes(mux *http.ServeMux, authMiddleware *AuthMiddleware) {
    // Protect document creation endpoint
    mux.Handle("/api/documents",
        authMiddleware.RequirePermission("dms", "document", "create")(
            http.HandlerFunc(createDocumentHandler),
        ),
    )

    // Protect employee management endpoint
    mux.Handle("/api/employees",
        authMiddleware.RequirePermission("hr", "employee", "manage")(
            http.HandlerFunc(manageEmployeeHandler),
        ),
    )

    // Protect division viewing endpoint
    mux.Handle("/api/divisions",
        authMiddleware.RequirePermission("organization", "division", "view")(
            http.HandlerFunc(listDivisionsHandler),
        ),
    )
}

func createDocumentHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Document created"))
}

func manageEmployeeHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Employee managed"))
}

func listDivisionsHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Divisions listed"))
}
```

### Context-Aware Permission Checking

```go
// CheckContextualPermission checks permission based on request context
func (m *AuthMiddleware) CheckContextualPermission(
    r *http.Request,
    namespace, resource, action string,
) (bool, error) {
    ctx := r.Context()

    // Extract entity and org context from request
    entityID := ctx.Value("entity_id").(string)
    divisionID := ctx.Value("division_id").(string)
    branchID := ctx.Value("branch_id").(string)
    departmentID := ctx.Value("department_id").(string)

    // Check with full organizational context
    resp, err := m.permissionClient.CheckPermissionWithOrgContext(
        ctx,
        entityID,
        namespace,
        resource,
        action,
        divisionID,
        branchID,
        departmentID,
    )

    if err != nil {
        return false, err
    }

    return resp.Effect == permissionv1.Effect_GRANT, nil
}
```

## Error Handling

### Comprehensive Error Handler

```go
// HandlePermissionError provides detailed error handling
func HandlePermissionError(err error, operation string) {
    if err == nil {
        return
    }

    var connectErr *connect.Error
    if errors.As(err, &connectErr) {
        switch connectErr.Code() {
        case connect.CodePermissionDenied:
            log.Printf("[%s] Access denied: %v", operation, connectErr.Message())
        case connect.CodeUnauthenticated:
            log.Printf("[%s] Authentication required: %v", operation, connectErr.Message())
        case connect.CodeNotFound:
            log.Printf("[%s] Subject or resource not found: %v", operation, connectErr.Message())
        case connect.CodeInvalidArgument:
            log.Printf("[%s] Invalid permission check parameters: %v", operation, connectErr.Message())
        default:
            log.Printf("[%s] Permission check error: %v (code: %v)",
                operation, connectErr.Message(), connectErr.Code())
        }
    } else {
        log.Printf("[%s] Unexpected error: %v", operation, err)
    }
}
```

## Best Practices

1. **Use Entity-Based Checks**: Always check permissions using entity IDs, not user IDs
2. **Include Organizational Context**: Provide division/branch/department for scoped checks
3. **Enable Inheritance**: Use WITH_INHERITANCE mode for hierarchical permissions
4. **Check Resource Ownership**: Verify ownership for resource-specific operations
5. **Analyze Resolution Path**: Use resolution path for debugging permission issues
6. **Cache Permission Results**: Cache frequently checked permissions with TTL
7. **Handle Temporal Permissions**: Always check valid_from and valid_until
8. **Fail Secure**: Default to deny if permission check fails
9. **Log Permission Checks**: Audit all permission checks for security
10. **Use Middleware**: Implement permission checking in middleware for consistency
11. **Batch Checks**: Check multiple permissions at once when possible
12. **Monitor Performance**: Track permission check latency and optimize
