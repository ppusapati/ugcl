# Creating Division, Branch, and Department - Complete Guide

This guide provides comprehensive, production-ready examples for creating and managing organizational hierarchy in the UGCL Backend v2 system. All examples include error handling, validation, and real-world scenarios.

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Client Setup](#client-setup)
4. [Creating Divisions](#creating-divisions)
5. [Creating Branches](#creating-branches)
6. [Creating Departments](#creating-departments)
7. [Querying Hierarchy](#querying-hierarchy)
8. [Updating Units](#updating-units)
9. [Soft Deleting](#soft-deleting)
10. [Complete Workflow Example](#complete-workflow-example)
11. [Error Handling](#error-handling)
12. [Best Practices](#best-practices)

## Overview

The organizational hierarchy in UGCL follows this structure:
- **Division** (Top level): Logical separation of business units
- **Branch** (Second level): Physical locations under divisions
- **Department** (Third level): Functional units that can be at business-level or division-specific

## Prerequisites

```go
import (
    "context"
    "fmt"
    "log"
    "time"

    "connectrpc.com/connect"
    organizationv1 "p9e.in/ugcl/organization/api/v1/organization"
    "p9e.in/ugcl/organization/api/v1/organization/organizationconnect"
    "google.golang.org/protobuf/types/known/structpb"
    "google.golang.org/protobuf/types/known/timestamppb"
    "google.golang.org/protobuf/types/known/wrapperspb"
)
```

## Client Setup

### Basic Client Configuration

```go
package main

import (
    "crypto/tls"
    "net/http"
    "time"
)

// OrganizationClient encapsulates the gRPC client configuration
type OrganizationClient struct {
    client organizationconnect.OrganizationServiceClient
    tenantID string
    userID   string
}

// NewOrganizationClient creates a new organization service client
func NewOrganizationClient(baseURL, tenantID, userID, authToken string) *OrganizationClient {
    // Configure HTTP client with timeout and TLS
    httpClient := &http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                MinVersion: tls.VersionTLS12,
            },
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 100,
            IdleConnTimeout:     90 * time.Second,
        },
    }

    // Create Connect client with authentication interceptor
    client := organizationconnect.NewOrganizationServiceClient(
        httpClient,
        baseURL,
        connect.WithInterceptors(
            authInterceptor(authToken),
        ),
    )

    return &OrganizationClient{
        client:   client,
        tenantID: tenantID,
        userID:   userID,
    }
}

// authInterceptor adds authentication token to requests
func authInterceptor(token string) connect.UnaryInterceptorFunc {
    return func(next connect.UnaryFunc) connect.UnaryFunc {
        return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
            req.Header().Set("Authorization", "Bearer "+token)
            return next(ctx, req)
        }
    }
}
```

### Client Usage Example

```go
func main() {
    // Initialize client
    client := NewOrganizationClient(
        "https://api.ugcl.example.com",
        "tenant-123",
        "user-456",
        "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    )

    ctx := context.Background()

    // Use client for operations
    division, err := client.CreateDivision(ctx, "COAL", "Coal Division", "Coal mining operations")
    if err != nil {
        log.Fatalf("Failed to create division: %v", err)
    }

    log.Printf("Created division: %s (ID: %s)", division.Name, division.Id)
}
```

## Creating Divisions

### Basic Division Creation

```go
// CreateDivision creates a new division with basic information
func (c *OrganizationClient) CreateDivision(
    ctx context.Context,
    code, name, description string,
) (*organizationv1.Division, error) {
    req := &organizationv1.CreateDivisionRequest{
        TenantId:    c.tenantID,
        Code:        code,
        Name:        name,
        Description: description,
        IsActive:    true,
        DisplayOrder: 1,
        CreatedBy:   c.userID,
    }

    resp, err := c.client.CreateDivision(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("create division failed: %w", err)
    }

    return resp.Msg, nil
}
```

### Division with Complete Details

```go
// CreateDivisionWithDetails creates a division with full configuration
func (c *OrganizationClient) CreateDivisionWithDetails(
    ctx context.Context,
    code, name, description, headUserID string,
    displayOrder int32,
    metadata map[string]interface{},
) (*organizationv1.Division, error) {
    // Convert metadata to protobuf Struct
    metadataStruct, err := structpb.NewStruct(metadata)
    if err != nil {
        return nil, fmt.Errorf("invalid metadata: %w", err)
    }

    req := &organizationv1.CreateDivisionRequest{
        TenantId:     c.tenantID,
        Code:         code,
        Name:         name,
        Description:  description,
        HeadUserId:   headUserID,
        IsActive:     true,
        DisplayOrder: displayOrder,
        Metadata:     metadataStruct,
        CreatedBy:    c.userID,
    }

    resp, err := c.client.CreateDivision(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("create division with details failed: %w", err)
    }

    return resp.Msg, nil
}
```

### Division Creation Example with Validation

```go
// CreateDivisionValidated creates a division with input validation
func (c *OrganizationClient) CreateDivisionValidated(
    ctx context.Context,
    code, name, description, headUserID string,
) (*organizationv1.Division, error) {
    // Validate inputs
    if code == "" {
        return nil, fmt.Errorf("division code is required")
    }
    if len(code) > 20 {
        return nil, fmt.Errorf("division code must be 20 characters or less")
    }
    if name == "" {
        return nil, fmt.Errorf("division name is required")
    }
    if len(name) > 100 {
        return nil, fmt.Errorf("division name must be 100 characters or less")
    }

    // Check if code already exists
    existing, err := c.GetDivisionByCode(ctx, code)
    if err == nil && existing != nil {
        return nil, fmt.Errorf("division with code %s already exists", code)
    }

    // Create metadata with creation context
    metadata := map[string]interface{}{
        "created_via": "api",
        "source":      "organizational_setup",
        "version":     "v2",
    }

    return c.CreateDivisionWithDetails(ctx, code, name, description, headUserID, 1, metadata)
}
```

## Creating Branches

### Basic Branch Creation

```go
// CreateBranch creates a new branch under a division
func (c *OrganizationClient) CreateBranch(
    ctx context.Context,
    divisionID, code, name, branchType string,
    address *organizationv1.Address,
    location *organizationv1.Location,
    contact *organizationv1.ContactInfo,
) (*organizationv1.Branch, error) {
    req := &organizationv1.CreateBranchRequest{
        DivisionId:  divisionID,
        TenantId:    c.tenantID,
        Code:        code,
        Name:        name,
        BranchType:  branchType,
        Address:     address,
        Location:    location,
        Contact:     contact,
        IsActive:    true,
        DisplayOrder: 1,
        CreatedBy:   c.userID,
    }

    resp, err := c.client.CreateBranch(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("create branch failed: %w", err)
    }

    return resp.Msg, nil
}
```

### Branch with Full Address and Location

```go
// CreateBranchComplete creates a branch with complete address and location details
func (c *OrganizationClient) CreateBranchComplete(
    ctx context.Context,
    divisionID, code, name, branchType, managerUserID string,
    addressLine1, addressLine2, city, state, country, postalCode string,
    latitude, longitude float64,
    phone, email string,
    displayOrder int32,
) (*organizationv1.Branch, error) {
    // Validate division exists
    division, err := c.client.GetDivision(ctx, connect.NewRequest(&organizationv1.GetDivisionRequest{
        Id: divisionID,
    }))
    if err != nil {
        return nil, fmt.Errorf("division %s not found: %w", divisionID, err)
    }
    if !division.Msg.IsActive {
        return nil, fmt.Errorf("cannot create branch under inactive division")
    }

    // Build address
    address := &organizationv1.Address{
        Line1:      addressLine1,
        Line2:      addressLine2,
        City:       city,
        State:      state,
        Country:    country,
        PostalCode: postalCode,
    }

    // Build location (GPS coordinates)
    location := &organizationv1.Location{
        Latitude:  latitude,
        Longitude: longitude,
    }

    // Build contact info
    contact := &organizationv1.ContactInfo{
        Phone: phone,
        Email: email,
    }

    // Create metadata
    metadata := map[string]interface{}{
        "timezone":        "Asia/Kolkata",
        "working_hours":   "09:00-18:00",
        "working_days":    []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat"},
        "established_year": time.Now().Year(),
    }
    metadataStruct, _ := structpb.NewStruct(metadata)

    req := &organizationv1.CreateBranchRequest{
        DivisionId:          divisionID,
        TenantId:            c.tenantID,
        Code:                code,
        Name:                name,
        BranchType:          branchType,
        Address:             address,
        Location:            location,
        Contact:             contact,
        BranchManagerUserId: managerUserID,
        IsActive:            true,
        DisplayOrder:        displayOrder,
        Metadata:            metadataStruct,
        CreatedBy:           c.userID,
    }

    resp, err := c.client.CreateBranch(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("create branch complete failed: %w", err)
    }

    return resp.Msg, nil
}
```

### Multiple Branches Creation

```go
// BranchInfo holds branch creation information
type BranchInfo struct {
    Code      string
    Name      string
    Type      string
    City      string
    State     string
    Latitude  float64
    Longitude float64
    Phone     string
    Email     string
}

// CreateMultipleBranches creates multiple branches for a division
func (c *OrganizationClient) CreateMultipleBranches(
    ctx context.Context,
    divisionID string,
    branches []BranchInfo,
) ([]*organizationv1.Branch, []error) {
    results := make([]*organizationv1.Branch, len(branches))
    errors := make([]error, len(branches))

    for i, info := range branches {
        address := &organizationv1.Address{
            Line1:   fmt.Sprintf("Branch Office - %s", info.Name),
            City:    info.City,
            State:   info.State,
            Country: "India",
        }

        location := &organizationv1.Location{
            Latitude:  info.Latitude,
            Longitude: info.Longitude,
        }

        contact := &organizationv1.ContactInfo{
            Phone: info.Phone,
            Email: info.Email,
        }

        branch, err := c.CreateBranch(
            ctx,
            divisionID,
            info.Code,
            info.Name,
            info.Type,
            address,
            location,
            contact,
        )

        results[i] = branch
        errors[i] = err

        if err != nil {
            log.Printf("Failed to create branch %s: %v", info.Code, err)
        } else {
            log.Printf("Created branch %s (ID: %s)", branch.Name, branch.Id)
        }

        // Add small delay to avoid rate limiting
        time.Sleep(100 * time.Millisecond)
    }

    return results, errors
}
```

## Creating Departments

### Business-Level Department

```go
// CreateBusinessDepartment creates a department at business level (no division)
func (c *OrganizationClient) CreateBusinessDepartment(
    ctx context.Context,
    code, name, description, departmentType, headUserID string,
) (*organizationv1.Department, error) {
    req := &organizationv1.CreateDepartmentRequest{
        TenantId:       c.tenantID,
        DivisionId:     "", // Empty for business-level
        Code:           code,
        Name:           name,
        Description:    description,
        DepartmentType: departmentType,
        HeadUserId:     headUserID,
        IsActive:       true,
        DisplayOrder:   1,
        CreatedBy:      c.userID,
    }

    resp, err := c.client.CreateDepartment(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("create business department failed: %w", err)
    }

    return resp.Msg, nil
}
```

### Division-Specific Department

```go
// CreateDivisionDepartment creates a department under a specific division
func (c *OrganizationClient) CreateDivisionDepartment(
    ctx context.Context,
    divisionID, code, name, description, departmentType, headUserID string,
) (*organizationv1.Department, error) {
    // Validate division exists
    _, err := c.client.GetDivision(ctx, connect.NewRequest(&organizationv1.GetDivisionRequest{
        Id: divisionID,
    }))
    if err != nil {
        return nil, fmt.Errorf("division not found: %w", err)
    }

    req := &organizationv1.CreateDepartmentRequest{
        TenantId:       c.tenantID,
        DivisionId:     divisionID,
        Code:           code,
        Name:           name,
        Description:    description,
        DepartmentType: departmentType,
        HeadUserId:     headUserID,
        IsActive:       true,
        DisplayOrder:   1,
        CreatedBy:      c.userID,
    }

    resp, err := c.client.CreateDepartment(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("create division department failed: %w", err)
    }

    return resp.Msg, nil
}
```

### Department with Hierarchy (Parent-Child)

```go
// CreateSubDepartment creates a sub-department under a parent department
func (c *OrganizationClient) CreateSubDepartment(
    ctx context.Context,
    parentDepartmentID, code, name, description, departmentType, headUserID string,
) (*organizationv1.Department, error) {
    // Get parent department to inherit division
    parent, err := c.client.GetDepartment(ctx, connect.NewRequest(&organizationv1.GetDepartmentRequest{
        Id: parentDepartmentID,
    }))
    if err != nil {
        return nil, fmt.Errorf("parent department not found: %w", err)
    }

    req := &organizationv1.CreateDepartmentRequest{
        TenantId:           c.tenantID,
        DivisionId:         parent.Msg.DivisionId, // Inherit from parent
        Code:               code,
        Name:               name,
        Description:        description,
        DepartmentType:     departmentType,
        HeadUserId:         headUserID,
        ParentDepartmentId: parentDepartmentID,
        IsActive:           true,
        DisplayOrder:       1,
        CreatedBy:          c.userID,
    }

    resp, err := c.client.CreateDepartment(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("create sub-department failed: %w", err)
    }

    return resp.Msg, nil
}
```

### Creating Department Hierarchy

```go
// CreateDepartmentHierarchy creates a complete department hierarchy
func (c *OrganizationClient) CreateDepartmentHierarchy(
    ctx context.Context,
    divisionID string,
) error {
    // Create HR department (parent)
    hrDept, err := c.CreateDivisionDepartment(
        ctx,
        divisionID,
        "HR",
        "Human Resources",
        "Manages all HR functions",
        "ADMINISTRATIVE",
        "hr-head-user-id",
    )
    if err != nil {
        return fmt.Errorf("failed to create HR department: %w", err)
    }
    log.Printf("Created HR Department: %s", hrDept.Id)

    // Create sub-departments under HR
    recruitmentDept, err := c.CreateSubDepartment(
        ctx,
        hrDept.Id,
        "HR-REC",
        "Recruitment",
        "Handles recruitment and onboarding",
        "OPERATIONAL",
        "recruitment-head-user-id",
    )
    if err != nil {
        return fmt.Errorf("failed to create recruitment department: %w", err)
    }
    log.Printf("Created Recruitment Department: %s", recruitmentDept.Id)

    payrollDept, err := c.CreateSubDepartment(
        ctx,
        hrDept.Id,
        "HR-PAY",
        "Payroll",
        "Manages payroll and compensation",
        "OPERATIONAL",
        "payroll-head-user-id",
    )
    if err != nil {
        return fmt.Errorf("failed to create payroll department: %w", err)
    }
    log.Printf("Created Payroll Department: %s", payrollDept.Id)

    return nil
}
```

## Querying Hierarchy

### Get Division by ID

```go
// GetDivision retrieves a division by ID
func (c *OrganizationClient) GetDivision(ctx context.Context, divisionID string) (*organizationv1.Division, error) {
    req := &organizationv1.GetDivisionRequest{
        Id: divisionID,
    }

    resp, err := c.client.GetDivision(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("get division failed: %w", err)
    }

    return resp.Msg, nil
}
```

### Get Division by Code

```go
// GetDivisionByCode retrieves a division by its unique code
func (c *OrganizationClient) GetDivisionByCode(ctx context.Context, code string) (*organizationv1.Division, error) {
    req := &organizationv1.GetDivisionByCodeRequest{
        TenantId: c.tenantID,
        Code:     code,
    }

    resp, err := c.client.GetDivisionByCode(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("get division by code failed: %w", err)
    }

    return resp.Msg, nil
}
```

### List All Divisions with Summary

```go
// ListDivisionSummaries retrieves all division summaries with counts
func (c *OrganizationClient) ListDivisionSummaries(ctx context.Context, activeOnly bool) ([]*organizationv1.DivisionSummary, error) {
    req := &organizationv1.ListDivisionSummariesRequest{
        TenantId: c.tenantID,
    }
    if activeOnly {
        req.IsActive = wrapperspb.Bool(true)
    }

    resp, err := c.client.ListDivisionSummaries(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("list division summaries failed: %w", err)
    }

    return resp.Msg.Summaries, nil
}
```

### List Branches by Division

```go
// ListBranchesByDivision retrieves all branches under a division
func (c *OrganizationClient) ListBranchesByDivision(
    ctx context.Context,
    divisionID string,
    activeOnly bool,
) ([]*organizationv1.Branch, error) {
    req := &organizationv1.ListBranchesByDivisionRequest{
        DivisionId: divisionID,
    }
    if activeOnly {
        req.IsActive = wrapperspb.Bool(true)
    }

    resp, err := c.client.ListBranchesByDivision(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("list branches by division failed: %w", err)
    }

    return resp.Msg.Branches, nil
}
```

### Get Complete Hierarchy

```go
// GetCompleteHierarchy retrieves the complete organizational hierarchy
func (c *OrganizationClient) GetCompleteHierarchy(ctx context.Context) (map[string]interface{}, error) {
    hierarchy := make(map[string]interface{})

    // Get all divisions
    divisionResp, err := c.client.ListDivisions(ctx, connect.NewRequest(&organizationv1.ListDivisionsRequest{
        TenantId: c.tenantID,
        IsActive: wrapperspb.Bool(true),
        Limit:    100,
        Offset:   0,
    }))
    if err != nil {
        return nil, fmt.Errorf("failed to list divisions: %w", err)
    }

    divisions := make([]map[string]interface{}, 0)
    for _, division := range divisionResp.Msg.Divisions {
        divisionData := map[string]interface{}{
            "id":   division.Id,
            "code": division.Code,
            "name": division.Name,
        }

        // Get branches for this division
        branchResp, err := c.client.ListBranchesByDivision(ctx, connect.NewRequest(&organizationv1.ListBranchesByDivisionRequest{
            DivisionId: division.Id,
            IsActive:   wrapperspb.Bool(true),
        }))
        if err == nil {
            branches := make([]map[string]interface{}, 0)
            for _, branch := range branchResp.Msg.Branches {
                branches = append(branches, map[string]interface{}{
                    "id":   branch.Id,
                    "code": branch.Code,
                    "name": branch.Name,
                    "city": branch.Address.City,
                })
            }
            divisionData["branches"] = branches
        }

        // Get departments for this division
        deptResp, err := c.client.ListDepartmentsByDivision(ctx, connect.NewRequest(&organizationv1.ListDepartmentsByDivisionRequest{
            DivisionId: division.Id,
            IsActive:   wrapperspb.Bool(true),
        }))
        if err == nil {
            departments := make([]map[string]interface{}, 0)
            for _, dept := range deptResp.Msg.Departments {
                departments = append(departments, map[string]interface{}{
                    "id":   dept.Id,
                    "code": dept.Code,
                    "name": dept.Name,
                })
            }
            divisionData["departments"] = departments
        }

        divisions = append(divisions, divisionData)
    }

    hierarchy["divisions"] = divisions
    return hierarchy, nil
}
```

## Updating Units

### Update Division

```go
// UpdateDivision updates division details
func (c *OrganizationClient) UpdateDivision(
    ctx context.Context,
    divisionID string,
    name, description *string,
    isActive *bool,
) (*organizationv1.Division, error) {
    req := &organizationv1.UpdateDivisionRequest{
        Id:        divisionID,
        UpdatedBy: c.userID,
    }

    if name != nil {
        req.Name = wrapperspb.String(*name)
    }
    if description != nil {
        req.Description = wrapperspb.String(*description)
    }
    if isActive != nil {
        req.IsActive = wrapperspb.Bool(*isActive)
    }

    resp, err := c.client.UpdateDivision(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("update division failed: %w", err)
    }

    return resp.Msg, nil
}
```

### Update Branch

```go
// UpdateBranch updates branch details including address and contact
func (c *OrganizationClient) UpdateBranch(
    ctx context.Context,
    branchID string,
    name *string,
    address *organizationv1.Address,
    contact *organizationv1.ContactInfo,
    isActive *bool,
) (*organizationv1.Branch, error) {
    req := &organizationv1.UpdateBranchRequest{
        Id:        branchID,
        UpdatedBy: c.userID,
    }

    if name != nil {
        req.Name = wrapperspb.String(*name)
    }
    if address != nil {
        req.Address = address
    }
    if contact != nil {
        req.Contact = contact
    }
    if isActive != nil {
        req.IsActive = wrapperspb.Bool(*isActive)
    }

    resp, err := c.client.UpdateBranch(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("update branch failed: %w", err)
    }

    return resp.Msg, nil
}
```

### Update Department

```go
// UpdateDepartment updates department details
func (c *OrganizationClient) UpdateDepartment(
    ctx context.Context,
    departmentID string,
    name, description, headUserID *string,
    isActive *bool,
) (*organizationv1.Department, error) {
    req := &organizationv1.UpdateDepartmentRequest{
        Id:        departmentID,
        UpdatedBy: c.userID,
    }

    if name != nil {
        req.Name = wrapperspb.String(*name)
    }
    if description != nil {
        req.Description = wrapperspb.String(*description)
    }
    if headUserID != nil {
        req.HeadUserId = wrapperspb.String(*headUserID)
    }
    if isActive != nil {
        req.IsActive = wrapperspb.Bool(*isActive)
    }

    resp, err := c.client.UpdateDepartment(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("update department failed: %w", err)
    }

    return resp.Msg, nil
}
```

## Soft Deleting

### Delete Division with Dependency Check

```go
// DeleteDivision performs soft delete with dependency validation
func (c *OrganizationClient) DeleteDivision(ctx context.Context, divisionID string) error {
    // Check for branches
    branches, err := c.ListBranchesByDivision(ctx, divisionID, true)
    if err != nil {
        return fmt.Errorf("failed to check branches: %w", err)
    }
    if len(branches) > 0 {
        return fmt.Errorf("cannot delete division with %d active branches", len(branches))
    }

    // Check for departments
    deptResp, err := c.client.ListDepartmentsByDivision(ctx, connect.NewRequest(&organizationv1.ListDepartmentsByDivisionRequest{
        DivisionId: divisionID,
        IsActive:   wrapperspb.Bool(true),
    }))
    if err != nil {
        return fmt.Errorf("failed to check departments: %w", err)
    }
    if len(deptResp.Msg.Departments) > 0 {
        return fmt.Errorf("cannot delete division with %d active departments", len(deptResp.Msg.Departments))
    }

    // Perform soft delete
    req := &organizationv1.DeleteDivisionRequest{
        Id:        divisionID,
        DeletedBy: c.userID,
    }

    _, err = c.client.DeleteDivision(ctx, connect.NewRequest(req))
    if err != nil {
        return fmt.Errorf("delete division failed: %w", err)
    }

    return nil
}
```

## Complete Workflow Example

```go
// SetupOrganizationalHierarchy creates a complete organizational structure
func SetupOrganizationalHierarchy(ctx context.Context) error {
    client := NewOrganizationClient(
        "https://api.ugcl.example.com",
        "tenant-ugcl-001",
        "admin-user-123",
        "auth-token-here",
    )

    // Step 1: Create Coal Division
    coalDivision, err := client.CreateDivisionValidated(
        ctx,
        "COAL",
        "Coal Division",
        "Manages all coal mining operations",
        "coal-head-user-id",
    )
    if err != nil {
        return fmt.Errorf("failed to create coal division: %w", err)
    }
    log.Printf("✓ Created Coal Division: %s", coalDivision.Id)

    // Step 2: Create branches under Coal Division
    branches := []BranchInfo{
        {
            Code:      "COAL-KRB",
            Name:      "Korba Branch",
            Type:      "MINING",
            City:      "Korba",
            State:     "Chhattisgarh",
            Latitude:  22.3595,
            Longitude: 82.7501,
            Phone:     "+91-7759-222000",
            Email:     "korba@ugcl.com",
        },
        {
            Code:      "COAL-BSP",
            Name:      "Bilaspur Branch",
            Type:      "MINING",
            City:      "Bilaspur",
            State:     "Chhattisgarh",
            Latitude:  22.0797,
            Longitude: 82.1409,
            Phone:     "+91-7752-222000",
            Email:     "bilaspur@ugcl.com",
        },
    }

    createdBranches, branchErrors := client.CreateMultipleBranches(ctx, coalDivision.Id, branches)
    for i, branch := range createdBranches {
        if branchErrors[i] != nil {
            log.Printf("✗ Failed to create branch %s: %v", branches[i].Code, branchErrors[i])
        } else {
            log.Printf("✓ Created Branch: %s (%s)", branch.Name, branch.Id)
        }
    }

    // Step 3: Create division-specific departments
    miningDept, err := client.CreateDivisionDepartment(
        ctx,
        coalDivision.Id,
        "MINING",
        "Mining Operations",
        "Oversees all mining activities",
        "OPERATIONAL",
        "mining-head-user-id",
    )
    if err != nil {
        return fmt.Errorf("failed to create mining department: %w", err)
    }
    log.Printf("✓ Created Mining Department: %s", miningDept.Id)

    // Step 4: Create business-level departments (HR, Finance, etc.)
    hrDept, err := client.CreateBusinessDepartment(
        ctx,
        "HR",
        "Human Resources",
        "Corporate HR functions",
        "ADMINISTRATIVE",
        "hr-head-user-id",
    )
    if err != nil {
        return fmt.Errorf("failed to create HR department: %w", err)
    }
    log.Printf("✓ Created HR Department: %s", hrDept.Id)

    // Step 5: Create sub-departments
    err = client.CreateDepartmentHierarchy(ctx, coalDivision.Id)
    if err != nil {
        return fmt.Errorf("failed to create department hierarchy: %w", err)
    }

    // Step 6: Query complete hierarchy
    hierarchy, err := client.GetCompleteHierarchy(ctx)
    if err != nil {
        return fmt.Errorf("failed to get hierarchy: %w", err)
    }
    log.Printf("Complete Hierarchy: %+v", hierarchy)

    return nil
}
```

## Error Handling

### Comprehensive Error Handler

```go
// HandleOrganizationError provides detailed error handling
func HandleOrganizationError(err error) {
    if err == nil {
        return
    }

    var connectErr *connect.Error
    if errors.As(err, &connectErr) {
        switch connectErr.Code() {
        case connect.CodeNotFound:
            log.Printf("Resource not found: %v", connectErr.Message())
        case connect.CodeAlreadyExists:
            log.Printf("Resource already exists: %v", connectErr.Message())
        case connect.CodeInvalidArgument:
            log.Printf("Invalid input: %v", connectErr.Message())
        case connect.CodePermissionDenied:
            log.Printf("Permission denied: %v", connectErr.Message())
        case connect.CodeUnauthenticated:
            log.Printf("Authentication required: %v", connectErr.Message())
        case connect.CodeInternal:
            log.Printf("Internal server error: %v", connectErr.Message())
        default:
            log.Printf("Error: %v (code: %v)", connectErr.Message(), connectErr.Code())
        }
    } else {
        log.Printf("Unexpected error: %v", err)
    }
}
```

## Best Practices

1. **Always validate inputs before creation**
2. **Check for duplicates using code lookups**
3. **Use transactions when creating multiple related entities**
4. **Implement retry logic for transient failures**
5. **Log all operations for audit trails**
6. **Use soft deletes with dependency checks**
7. **Include metadata for tracking and debugging**
8. **Handle errors gracefully with proper error messages**
9. **Use batch operations for creating multiple entities**
10. **Implement proper authentication and authorization**
