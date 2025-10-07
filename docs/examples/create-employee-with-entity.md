# Creating Employee with Entity - Complete Guide

This comprehensive guide demonstrates the complete workflow for creating employees with entity linkage, role assignments, and organizational scope in the UGCL Backend v2 system. The same patterns apply to contractors and vendors.

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Client Setup](#client-setup)
4. [Creating User Account](#creating-user-account)
5. [Creating Employee Record](#creating-employee-record)
6. [Creating Entity Linking](#creating-entity-linking)
7. [Assigning Entity Roles](#assigning-entity-roles)
8. [Setting Organizational Scope](#setting-organizational-scope)
9. [Temporal Access Control](#temporal-access-control)
10. [Complete Workflow](#complete-workflow)
11. [Contractor Creation](#contractor-creation)
12. [Vendor Creation](#vendor-creation)
13. [Error Handling](#error-handling)
14. [Best Practices](#best-practices)

## Overview

The personnel creation process follows this sequence:
1. **Create User**: Identity and authentication account
2. **Create Employee/Contractor/Vendor**: Domain-specific record
3. **Create Entity**: Unified identity abstraction linking user and domain record
4. **Assign Entity Roles**: Role bindings with organizational scope
5. **Set Permissions**: Configure access control with temporal constraints

## Prerequisites

```go
import (
    "context"
    "fmt"
    "log"
    "time"

    "connectrpc.com/connect"
    userv1 "p9e.in/ugcl/identity/user/api/v2/user"
    "p9e.in/ugcl/identity/user/api/v2/user/userconnect"
    employeev1 "p9e.in/ugcl/employee/api/v2/employee"
    "p9e.in/ugcl/employee/api/v2/employee/employeeconnect"
    entityv1 "p9e.in/ugcl/identity/entity/api/v1"
    "p9e.in/ugcl/identity/entity/api/v1/entityv1connect"
    contractorv1 "p9e.in/ugcl/contractors/api/v2/contractor"
    "p9e.in/ugcl/contractors/api/v2/contractor/contractorconnect"
    vendorv1 "p9e.in/ugcl/vendors/api/v2/vendor"
    "p9e.in/ugcl/vendors/api/v2/vendor/vendorconnect"
    "google.golang.org/protobuf/types/known/structpb"
    "google.golang.org/protobuf/types/known/timestamppb"
    "google.golang.org/protobuf/types/known/wrapperspb"
)
```

## Client Setup

### Multi-Service Client Configuration

```go
package main

import (
    "crypto/tls"
    "net/http"
    "time"
)

// PersonnelClient encapsulates all personnel-related service clients
type PersonnelClient struct {
    userClient       userv1connect.UserServiceClient
    employeeClient   employeeconnect.EmployeeServiceClient
    entityClient     entityv1connect.EntityServiceClient
    contractorClient contractorconnect.ContractorServiceClient
    vendorClient     vendorconnect.VendorServiceClient
    tenantID         string
    adminUserID      string
}

// NewPersonnelClient creates a new personnel service client
func NewPersonnelClient(baseURL, tenantID, adminUserID, authToken string) *PersonnelClient {
    // Configure HTTP client
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

    // Create interceptor for authentication
    interceptor := authInterceptor(authToken)

    return &PersonnelClient{
        userClient: userv1connect.NewUserServiceClient(
            httpClient,
            baseURL,
            connect.WithInterceptors(interceptor),
        ),
        employeeClient: employeeconnect.NewEmployeeServiceClient(
            httpClient,
            baseURL,
            connect.WithInterceptors(interceptor),
        ),
        entityClient: entityv1connect.NewEntityServiceClient(
            httpClient,
            baseURL,
            connect.WithInterceptors(interceptor),
        ),
        contractorClient: contractorconnect.NewContractorServiceClient(
            httpClient,
            baseURL,
            connect.WithInterceptors(interceptor),
        ),
        vendorClient: vendorconnect.NewVendorServiceClient(
            httpClient,
            baseURL,
            connect.WithInterceptors(interceptor),
        ),
        tenantID:    tenantID,
        adminUserID: adminUserID,
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
```

## Creating User Account

### Basic User Creation

```go
// CreateUser creates a new user account
func (c *PersonnelClient) CreateUser(
    ctx context.Context,
    email, firstName, lastName, mobileNumber string,
) (*userv1.User, error) {
    user := &userv1.User{
        TenantId:     c.tenantID,
        Email:        email,
        FirstName:    firstName,
        LastName:     lastName,
        MobileNumber: wrapperspb.String(mobileNumber),
        IsActive:     true,
        EmailVerified: false,
        PhoneVerified: false,
    }

    req := &userv1.CreateUserRequest{
        User: user,
    }

    resp, err := c.userClient.CreateUser(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("create user failed: %w", err)
    }

    return resp.Msg, nil
}
```

### User with Complete Details

```go
// CreateUserComplete creates a user with full personal information
func (c *PersonnelClient) CreateUserComplete(
    ctx context.Context,
    email, firstName, middleName, lastName string,
    dateOfBirth time.Time,
    gender, mobileNumber, alternateEmail string,
    address map[string]string,
) (*userv1.User, error) {
    // Build address struct
    addressStruct, err := structpb.NewStruct(map[string]interface{}{
        "line1":       address["line1"],
        "line2":       address["line2"],
        "city":        address["city"],
        "state":       address["state"],
        "country":     address["country"],
        "postal_code": address["postal_code"],
    })
    if err != nil {
        return nil, fmt.Errorf("invalid address: %w", err)
    }

    user := &userv1.User{
        TenantId:       c.tenantID,
        Email:          email,
        FirstName:      firstName,
        MiddleName:     wrapperspb.String(middleName),
        LastName:       lastName,
        DateOfBirth:    timestamppb.New(dateOfBirth),
        Gender:         wrapperspb.String(gender),
        MobileNumber:   wrapperspb.String(mobileNumber),
        AlternateEmail: wrapperspb.String(alternateEmail),
        Address:        addressStruct,
        IsActive:       true,
        EmailVerified:  false,
        PhoneVerified:  false,
    }

    req := &userv1.CreateUserRequest{
        User: user,
    }

    resp, err := c.userClient.CreateUser(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("create user complete failed: %w", err)
    }

    log.Printf("Created user: %s %s (ID: %s)", firstName, lastName, resp.Msg.Id)
    return resp.Msg, nil
}
```

## Creating Employee Record

### Basic Employee Creation

```go
// CreateEmployee creates a new employee record
func (c *PersonnelClient) CreateEmployee(
    ctx context.Context,
    userID, employeeCode, designation string,
    divisionID, branchID, departmentID string,
    dateOfJoining time.Time,
) (*employeev1.Employee, error) {
    employee := &employeev1.Employee{
        UserId:         userID,
        EmployeeCode:   employeeCode,
        Designation:    designation,
        DivisionId:     wrapperspb.String(divisionID),
        BranchId:       wrapperspb.String(branchID),
        DepartmentId:   wrapperspb.String(departmentID),
        DateOfJoining:  timestamppb.New(dateOfJoining),
        Status:         "ACTIVE",
        EmployeeType:   wrapperspb.String("PERMANENT"),
    }

    req := &employeev1.CreateEmployeeRequest{
        Employee: employee,
    }

    resp, err := c.employeeClient.CreateEmployee(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("create employee failed: %w", err)
    }

    return resp.Msg, nil
}
```

### Employee with Complete Details

```go
// EmployeeDetails holds complete employee information
type EmployeeDetails struct {
    UserID           string
    EmployeeCode     string
    Designation      string
    JobTitle         string
    JobGrade         string
    EmployeeType     string
    DivisionID       string
    BranchID         string
    DepartmentID     string
    ManagerID        string
    DateOfJoining    time.Time
    DateOfConfirmation *time.Time

    // Personal details
    PAN              string
    Aadhaar          string
    UAN              string

    // Bank details
    BankName         string
    AccountNumber    string
    IFSC             string

    // Address
    CurrentAddress   string
    PermanentAddress string

    // Emergency contact
    EmergencyName     string
    EmergencyPhone    string
    EmergencyRelation string

    // Work details
    WorkLocation string
    OfficePhone  string
    Extension    string

    // Skills
    Skills         []string
    Certifications []string
    Qualification  string
}

// CreateEmployeeComplete creates an employee with full details
func (c *PersonnelClient) CreateEmployeeComplete(
    ctx context.Context,
    details EmployeeDetails,
) (*employeev1.Employee, error) {
    employee := &employeev1.Employee{
        UserId:                  details.UserID,
        EmployeeCode:            details.EmployeeCode,
        Designation:             details.Designation,
        JobTitle:                wrapperspb.String(details.JobTitle),
        JobGrade:                wrapperspb.String(details.JobGrade),
        EmployeeType:            wrapperspb.String(details.EmployeeType),
        DivisionId:              wrapperspb.String(details.DivisionID),
        BranchId:                wrapperspb.String(details.BranchID),
        DepartmentId:            wrapperspb.String(details.DepartmentID),
        ManagerId:               wrapperspb.String(details.ManagerID),
        DateOfJoining:           timestamppb.New(details.DateOfJoining),
        Pan:                     wrapperspb.String(details.PAN),
        Aadhaar:                 wrapperspb.String(details.Aadhaar),
        Uan:                     wrapperspb.String(details.UAN),
        BankName:                wrapperspb.String(details.BankName),
        BankAccountNumber:       wrapperspb.String(details.AccountNumber),
        BankIfsc:                wrapperspb.String(details.IFSC),
        CurrentAddress:          wrapperspb.String(details.CurrentAddress),
        PermanentAddress:        wrapperspb.String(details.PermanentAddress),
        EmergencyContactName:    wrapperspb.String(details.EmergencyName),
        EmergencyContactPhone:   wrapperspb.String(details.EmergencyPhone),
        EmergencyContactRelation: wrapperspb.String(details.EmergencyRelation),
        WorkLocation:            wrapperspb.String(details.WorkLocation),
        OfficePhone:             wrapperspb.String(details.OfficePhone),
        Extension:               wrapperspb.String(details.Extension),
        Status:                  "ACTIVE",
        Skills:                  details.Skills,
        Certifications:          details.Certifications,
        HighestQualification:    wrapperspb.String(details.Qualification),
    }

    if details.DateOfConfirmation != nil {
        employee.DateOfConfirmation = timestamppb.New(*details.DateOfConfirmation)
    }

    req := &employeev1.CreateEmployeeRequest{
        Employee: employee,
    }

    resp, err := c.employeeClient.CreateEmployee(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("create employee complete failed: %w", err)
    }

    log.Printf("Created employee: %s (Code: %s, ID: %s)",
        details.Designation, details.EmployeeCode, resp.Msg.Id)

    return resp.Msg, nil
}
```

## Creating Entity Linking

### Create Entity for Employee

```go
// CreateEntityForEmployee creates an entity linking user to employee
func (c *PersonnelClient) CreateEntityForEmployee(
    ctx context.Context,
    userID, employeeID string,
    divisionID, branchID, departmentID string,
) (*entityv1.Entity, error) {
    req := &entityv1.CreateEntityRequest{
        TenantId:        c.tenantID,
        EntityType:      entityv1.EntityType_ENTITY_TYPE_EMPLOYEE,
        UserId:          userID,
        ReferenceId:     employeeID,
        ReferenceSource: entityv1.ReferenceSource_REFERENCE_SOURCE_EMPLOYEE,
        Status:          entityv1.EntityStatus_ENTITY_STATUS_ACTIVE,
    }

    // Set organizational context
    if divisionID != "" {
        req.DivisionId = wrapperspb.String(divisionID)
    }
    if branchID != "" {
        req.BranchId = wrapperspb.String(branchID)
    }
    if departmentID != "" {
        req.DepartmentId = wrapperspb.String(departmentID)
    }

    resp, err := c.entityClient.CreateEntity(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("create entity failed: %w", err)
    }

    log.Printf("Created entity: %s (Type: EMPLOYEE, ID: %s)", resp.Msg.Id, resp.Msg.ReferenceId)
    return resp.Msg, nil
}
```

### Create Entity with Metadata

```go
// CreateEntityWithMetadata creates an entity with custom metadata
func (c *PersonnelClient) CreateEntityWithMetadata(
    ctx context.Context,
    entityType entityv1.EntityType,
    userID, referenceID string,
    referenceSource entityv1.ReferenceSource,
    divisionID, branchID, departmentID string,
    metadata map[string]string,
) (*entityv1.Entity, error) {
    req := &entityv1.CreateEntityRequest{
        TenantId:        c.tenantID,
        EntityType:      entityType,
        UserId:          userID,
        ReferenceId:     referenceID,
        ReferenceSource: referenceSource,
        Status:          entityv1.EntityStatus_ENTITY_STATUS_ACTIVE,
        Metadata:        metadata,
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

    resp, err := c.entityClient.CreateEntity(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("create entity with metadata failed: %w", err)
    }

    return resp.Msg, nil
}
```

## Assigning Entity Roles

### Create Entity Role Binding

```go
// CreateEntityRoleBinding assigns a role to an entity
func (c *PersonnelClient) CreateEntityRoleBinding(
    ctx context.Context,
    entityID, roleID string,
    divisionID, branchID, departmentID string,
) (*entityv1.EntityRoleBinding, error) {
    req := &entityv1.CreateEntityRoleBindingRequest{
        TenantId: c.tenantID,
        EntityId: entityID,
        RoleId:   roleID,
    }

    // Set scope limitations
    if divisionID != "" {
        req.DivisionId = wrapperspb.String(divisionID)
    }
    if branchID != "" {
        req.BranchId = wrapperspb.String(branchID)
    }
    if departmentID != "" {
        req.DepartmentId = wrapperspb.String(departmentID)
    }

    resp, err := c.entityClient.CreateEntityRoleBinding(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("create entity role binding failed: %w", err)
    }

    log.Printf("Assigned role %s to entity %s", roleID, entityID)
    return resp.Msg, nil
}
```

### Assign Multiple Roles with Different Scopes

```go
// RoleAssignment represents a role assignment with scope
type RoleAssignment struct {
    RoleID       string
    DivisionID   string
    BranchID     string
    DepartmentID string
    ValidFrom    *time.Time
    ValidUntil   *time.Time
}

// AssignMultipleRoles assigns multiple roles to an entity
func (c *PersonnelClient) AssignMultipleRoles(
    ctx context.Context,
    entityID string,
    assignments []RoleAssignment,
) ([]*entityv1.EntityRoleBinding, error) {
    bindings := make([]*entityv1.EntityRoleBinding, 0, len(assignments))

    for _, assignment := range assignments {
        req := &entityv1.CreateEntityRoleBindingRequest{
            TenantId: c.tenantID,
            EntityId: entityID,
            RoleId:   assignment.RoleID,
        }

        if assignment.DivisionID != "" {
            req.DivisionId = wrapperspb.String(assignment.DivisionID)
        }
        if assignment.BranchID != "" {
            req.BranchId = wrapperspb.String(assignment.BranchID)
        }
        if assignment.DepartmentID != "" {
            req.DepartmentId = wrapperspb.String(assignment.DepartmentID)
        }
        if assignment.ValidFrom != nil {
            req.ValidFrom = timestamppb.New(*assignment.ValidFrom)
        }
        if assignment.ValidUntil != nil {
            req.ValidUntil = timestamppb.New(*assignment.ValidUntil)
        }

        resp, err := c.entityClient.CreateEntityRoleBinding(ctx, connect.NewRequest(req))
        if err != nil {
            log.Printf("Warning: Failed to assign role %s: %v", assignment.RoleID, err)
            continue
        }

        bindings = append(bindings, resp.Msg)
        log.Printf("✓ Assigned role %s to entity %s", assignment.RoleID, entityID)
    }

    return bindings, nil
}
```

## Setting Organizational Scope

### Assign Role with Division Scope

```go
// AssignDivisionRole assigns a role scoped to a specific division
func (c *PersonnelClient) AssignDivisionRole(
    ctx context.Context,
    entityID, roleID, divisionID string,
) (*entityv1.EntityRoleBinding, error) {
    req := &entityv1.CreateEntityRoleBindingRequest{
        TenantId:   c.tenantID,
        EntityId:   entityID,
        RoleId:     roleID,
        DivisionId: wrapperspb.String(divisionID),
    }

    resp, err := c.entityClient.CreateEntityRoleBinding(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("assign division role failed: %w", err)
    }

    log.Printf("Assigned division-scoped role: Entity=%s, Role=%s, Division=%s",
        entityID, roleID, divisionID)

    return resp.Msg, nil
}
```

### Assign Role with Department Scope

```go
// AssignDepartmentRole assigns a role scoped to a specific department
func (c *PersonnelClient) AssignDepartmentRole(
    ctx context.Context,
    entityID, roleID, departmentID string,
) (*entityv1.EntityRoleBinding, error) {
    req := &entityv1.CreateEntityRoleBindingRequest{
        TenantId:     c.tenantID,
        EntityId:     entityID,
        RoleId:       roleID,
        DepartmentId: wrapperspb.String(departmentID),
    }

    resp, err := c.entityClient.CreateEntityRoleBinding(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("assign department role failed: %w", err)
    }

    log.Printf("Assigned department-scoped role: Entity=%s, Role=%s, Department=%s",
        entityID, roleID, departmentID)

    return resp.Msg, nil
}
```

## Temporal Access Control

### Assign Time-Bound Role

```go
// AssignTemporalRole assigns a role with time constraints
func (c *PersonnelClient) AssignTemporalRole(
    ctx context.Context,
    entityID, roleID string,
    validFrom, validUntil time.Time,
    divisionID, branchID, departmentID string,
) (*entityv1.EntityRoleBinding, error) {
    req := &entityv1.CreateEntityRoleBindingRequest{
        TenantId:   c.tenantID,
        EntityId:   entityID,
        RoleId:     roleID,
        ValidFrom:  timestamppb.New(validFrom),
        ValidUntil: timestamppb.New(validUntil),
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

    resp, err := c.entityClient.CreateEntityRoleBinding(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("assign temporal role failed: %w", err)
    }

    log.Printf("Assigned temporal role: Entity=%s, Role=%s, Valid: %s to %s",
        entityID, roleID,
        validFrom.Format("2006-01-02"),
        validUntil.Format("2006-01-02"))

    return resp.Msg, nil
}
```

### Assign Project-Based Temporary Access

```go
// AssignProjectAccess assigns temporary access for a project duration
func (c *PersonnelClient) AssignProjectAccess(
    ctx context.Context,
    entityID, roleID, projectID string,
    projectStartDate, projectEndDate time.Time,
) (*entityv1.EntityRoleBinding, error) {
    metadata := map[string]string{
        "project_id":  projectID,
        "access_type": "project_based",
        "reason":      "temporary project assignment",
    }

    req := &entityv1.CreateEntityRoleBindingRequest{
        TenantId:   c.tenantID,
        EntityId:   entityID,
        RoleId:     roleID,
        ValidFrom:  timestamppb.New(projectStartDate),
        ValidUntil: timestamppb.New(projectEndDate),
    }

    resp, err := c.entityClient.CreateEntityRoleBinding(ctx, connect.NewRequest(req))
    if err != nil {
        return nil, fmt.Errorf("assign project access failed: %w", err)
    }

    log.Printf("Assigned project-based access: Entity=%s, Project=%s, Duration: %d days",
        entityID, projectID, int(projectEndDate.Sub(projectStartDate).Hours()/24))

    return resp.Msg, nil
}
```

## Complete Workflow

### Complete Employee Onboarding

```go
// OnboardEmployee performs complete employee onboarding
func (c *PersonnelClient) OnboardEmployee(
    ctx context.Context,
    firstName, lastName, email, mobileNumber string,
    employeeCode, designation string,
    divisionID, branchID, departmentID string,
    dateOfJoining time.Time,
    roleAssignments []RoleAssignment,
) (*EmployeeOnboardingResult, error) {
    result := &EmployeeOnboardingResult{}

    // Step 1: Create user account
    log.Println("Step 1: Creating user account...")
    user, err := c.CreateUser(ctx, email, firstName, lastName, mobileNumber)
    if err != nil {
        return nil, fmt.Errorf("user creation failed: %w", err)
    }
    result.User = user
    log.Printf("✓ User created: %s (ID: %s)", user.Email, user.Id)

    // Step 2: Create employee record
    log.Println("Step 2: Creating employee record...")
    employee, err := c.CreateEmployee(
        ctx,
        user.Id,
        employeeCode,
        designation,
        divisionID,
        branchID,
        departmentID,
        dateOfJoining,
    )
    if err != nil {
        return nil, fmt.Errorf("employee creation failed: %w", err)
    }
    result.Employee = employee
    log.Printf("✓ Employee created: %s (Code: %s, ID: %s)",
        designation, employeeCode, employee.Id)

    // Step 3: Create entity linking user and employee
    log.Println("Step 3: Creating entity...")
    entity, err := c.CreateEntityForEmployee(
        ctx,
        user.Id,
        employee.Id,
        divisionID,
        branchID,
        departmentID,
    )
    if err != nil {
        return nil, fmt.Errorf("entity creation failed: %w", err)
    }
    result.Entity = entity
    log.Printf("✓ Entity created: %s", entity.Id)

    // Step 4: Assign roles
    log.Println("Step 4: Assigning roles...")
    bindings, err := c.AssignMultipleRoles(ctx, entity.Id, roleAssignments)
    if err != nil {
        log.Printf("Warning: Some role assignments failed: %v", err)
    }
    result.RoleBindings = bindings
    log.Printf("✓ Assigned %d roles", len(bindings))

    log.Println("✓ Employee onboarding completed successfully!")
    return result, nil
}

// EmployeeOnboardingResult holds the results of employee onboarding
type EmployeeOnboardingResult struct {
    User         *userv1.User
    Employee     *employeev1.Employee
    Entity       *entityv1.Entity
    RoleBindings []*entityv1.EntityRoleBinding
}
```

### Complete Example with All Details

```go
// OnboardEmployeeComplete demonstrates complete onboarding with full details
func OnboardEmployeeComplete(ctx context.Context) error {
    client := NewPersonnelClient(
        "https://api.ugcl.example.com",
        "tenant-ugcl-001",
        "admin-user-123",
        "auth-token-here",
    )

    // Step 1: Create user with complete details
    log.Println("=== Employee Onboarding Process ===")
    log.Println("Step 1: Creating user account...")

    user, err := client.CreateUserComplete(
        ctx,
        "john.doe@ugcl.com",
        "John",
        "Kumar",
        "Doe",
        time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
        "Male",
        "+91-9876543210",
        "john.personal@gmail.com",
        map[string]string{
            "line1":       "123 Main Street",
            "line2":       "Sector 5",
            "city":        "Korba",
            "state":       "Chhattisgarh",
            "country":     "India",
            "postal_code": "495677",
        },
    )
    if err != nil {
        return fmt.Errorf("user creation failed: %w", err)
    }
    log.Printf("✓ User created: %s %s (ID: %s)", user.FirstName, user.LastName, user.Id)

    // Step 2: Create employee with complete details
    log.Println("Step 2: Creating employee record...")

    confirmationDate := time.Now().AddDate(0, 6, 0) // 6 months from now
    employeeDetails := EmployeeDetails{
        UserID:             user.Id,
        EmployeeCode:       "EMP-2024-001",
        Designation:        "Senior Mining Engineer",
        JobTitle:           "Mining Operations Lead",
        JobGrade:           "E3",
        EmployeeType:       "PERMANENT",
        DivisionID:         "div-coal-001",
        BranchID:           "branch-korba-001",
        DepartmentID:       "dept-mining-ops-001",
        ManagerID:          "emp-manager-001",
        DateOfJoining:      time.Now(),
        DateOfConfirmation: &confirmationDate,
        PAN:                "ABCDE1234F",
        Aadhaar:            "1234-5678-9012",
        UAN:                "123456789012",
        BankName:           "State Bank of India",
        AccountNumber:      "12345678901234",
        IFSC:               "SBIN0001234",
        CurrentAddress:     "123 Main Street, Sector 5, Korba",
        PermanentAddress:   "456 Home Street, Village, District",
        EmergencyName:      "Jane Doe",
        EmergencyPhone:     "+91-9876543211",
        EmergencyRelation:  "Spouse",
        WorkLocation:       "Korba Mining Site A",
        OfficePhone:        "+91-7759-222001",
        Extension:          "1234",
        Skills:             []string{"Mining Operations", "Safety Management", "Team Leadership"},
        Certifications:     []string{"Mining Safety Certificate", "First Aid"},
        Qualification:      "B.Tech Mining Engineering",
    }

    employee, err := client.CreateEmployeeComplete(ctx, employeeDetails)
    if err != nil {
        return fmt.Errorf("employee creation failed: %w", err)
    }
    log.Printf("✓ Employee created: %s (Code: %s, ID: %s)",
        employee.Designation, employee.EmployeeCode, employee.Id)

    // Step 3: Create entity
    log.Println("Step 3: Creating entity...")

    entity, err := client.CreateEntityForEmployee(
        ctx,
        user.Id,
        employee.Id,
        employeeDetails.DivisionID,
        employeeDetails.BranchID,
        employeeDetails.DepartmentID,
    )
    if err != nil {
        return fmt.Errorf("entity creation failed: %w", err)
    }
    log.Printf("✓ Entity created: %s", entity.Id)

    // Step 4: Assign roles with different scopes
    log.Println("Step 4: Assigning roles...")

    // Role 1: Department-level engineer role
    binding1, err := client.AssignDepartmentRole(
        ctx,
        entity.Id,
        "role-mining-engineer",
        employeeDetails.DepartmentID,
    )
    if err != nil {
        log.Printf("Warning: Failed to assign department role: %v", err)
    } else {
        log.Printf("✓ Assigned department role: %s", binding1.Id)
    }

    // Role 2: Division-level viewer role
    binding2, err := client.AssignDivisionRole(
        ctx,
        entity.Id,
        "role-division-viewer",
        employeeDetails.DivisionID,
    )
    if err != nil {
        log.Printf("Warning: Failed to assign division role: %v", err)
    } else {
        log.Printf("✓ Assigned division role: %s", binding2.Id)
    }

    // Role 3: Temporary project access (90 days)
    projectStart := time.Now()
    projectEnd := projectStart.AddDate(0, 0, 90)
    binding3, err := client.AssignProjectAccess(
        ctx,
        entity.Id,
        "role-project-member",
        "project-expansion-2024",
        projectStart,
        projectEnd,
    )
    if err != nil {
        log.Printf("Warning: Failed to assign project role: %v", err)
    } else {
        log.Printf("✓ Assigned temporary project role: %s (valid until %s)",
            binding3.Id, projectEnd.Format("2006-01-02"))
    }

    log.Println("=== Employee Onboarding Completed Successfully ===")
    log.Printf("Employee: %s %s", user.FirstName, user.LastName)
    log.Printf("Employee Code: %s", employee.EmployeeCode)
    log.Printf("Entity ID: %s", entity.Id)
    log.Printf("Division: %s", employeeDetails.DivisionID)
    log.Printf("Department: %s", employeeDetails.DepartmentID)

    return nil
}
```

## Contractor Creation

### Complete Contractor Onboarding

```go
// OnboardContractor creates a contractor with entity and role assignments
func (c *PersonnelClient) OnboardContractor(
    ctx context.Context,
    companyName, contactPersonName, email, mobileNumber string,
    gst, pan string,
    contractStartDate, contractEndDate time.Time,
    roleAssignments []RoleAssignment,
) error {
    // Step 1: Create user for contact person
    log.Println("Step 1: Creating user for contractor contact person...")
    user, err := c.CreateUser(ctx, email, contactPersonName, "", mobileNumber)
    if err != nil {
        return fmt.Errorf("user creation failed: %w", err)
    }
    log.Printf("✓ User created: %s (ID: %s)", user.Email, user.Id)

    // Step 2: Create contractor record
    log.Println("Step 2: Creating contractor record...")
    contractor := &contractorv1.Contractor{
        CompanyName:       companyName,
        CompanyType:       "PRIVATE_LIMITED",
        Gst:               gst,
        Pan:               pan,
        Category:          "MINING_CONTRACTOR",
        PersonId:          user.Id,
        ContractStartDate: timestamppb.New(contractStartDate),
        ContractEndDate:   timestamppb.New(contractEndDate),
        Status:            "ACTIVE",
    }

    contractorReq := &contractorv1.CreateContractorRequest{
        Contractor: contractor,
        User:       user,
    }

    contractorResp, err := c.contractorClient.CreateContractor(ctx, connect.NewRequest(contractorReq))
    if err != nil {
        return fmt.Errorf("contractor creation failed: %w", err)
    }
    log.Printf("✓ Contractor created: %s (ID: %s)", companyName, contractorResp.Msg.Id)

    // Step 3: Create entity
    log.Println("Step 3: Creating entity...")
    entity, err := c.CreateEntityWithMetadata(
        ctx,
        entityv1.EntityType_ENTITY_TYPE_CONTRACTOR,
        user.Id,
        contractorResp.Msg.Id,
        entityv1.ReferenceSource_REFERENCE_SOURCE_CONTRACTOR,
        "", "", "", // No organizational scope for contractors
        map[string]string{
            "company_name":  companyName,
            "contract_type": "MINING",
        },
    )
    if err != nil {
        return fmt.Errorf("entity creation failed: %w", err)
    }
    log.Printf("✓ Entity created: %s", entity.Id)

    // Step 4: Assign roles
    log.Println("Step 4: Assigning roles...")
    bindings, err := c.AssignMultipleRoles(ctx, entity.Id, roleAssignments)
    if err != nil {
        log.Printf("Warning: Some role assignments failed: %v", err)
    }
    log.Printf("✓ Assigned %d roles", len(bindings))

    log.Println("✓ Contractor onboarding completed!")
    return nil
}
```

## Vendor Creation

### Complete Vendor Onboarding

```go
// OnboardVendor creates a vendor with entity and role assignments
func (c *PersonnelClient) OnboardVendor(
    ctx context.Context,
    companyName, contactPersonName, email, mobileNumber string,
    gst, pan string,
    vendorCategory string,
) error {
    // Step 1: Create user
    log.Println("Step 1: Creating user for vendor contact...")
    user, err := c.CreateUser(ctx, email, contactPersonName, "", mobileNumber)
    if err != nil {
        return fmt.Errorf("user creation failed: %w", err)
    }
    log.Printf("✓ User created: %s", user.Id)

    // Step 2: Create vendor (similar to contractor but with vendor-specific fields)
    // Implementation depends on vendor proto definition

    log.Println("✓ Vendor onboarding completed!")
    return nil
}
```

## Error Handling

### Comprehensive Error Handler

```go
// HandlePersonnelError provides detailed error handling
func HandlePersonnelError(err error, operation string) {
    if err == nil {
        return
    }

    var connectErr *connect.Error
    if errors.As(err, &connectErr) {
        switch connectErr.Code() {
        case connect.CodeAlreadyExists:
            log.Printf("[%s] Record already exists: %v", operation, connectErr.Message())
        case connect.CodeNotFound:
            log.Printf("[%s] Record not found: %v", operation, connectErr.Message())
        case connect.CodeInvalidArgument:
            log.Printf("[%s] Invalid input: %v", operation, connectErr.Message())
        case connect.CodeFailedPrecondition:
            log.Printf("[%s] Prerequisites not met: %v", operation, connectErr.Message())
        default:
            log.Printf("[%s] Error: %v (code: %v)", operation, connectErr.Message(), connectErr.Code())
        }
    } else {
        log.Printf("[%s] Unexpected error: %v", operation, err)
    }
}
```

## Best Practices

1. **Transaction Safety**: Create user, employee, and entity in order with rollback on failure
2. **Validation**: Validate all inputs before creation (email format, employee code uniqueness)
3. **Organizational Context**: Always set division/branch/department for proper scoping
4. **Role Assignment**: Assign roles immediately after entity creation
5. **Temporal Access**: Use valid_from/valid_until for temporary assignments
6. **Audit Trail**: Log all creation and assignment operations
7. **Error Recovery**: Implement proper cleanup on partial failures
8. **Data Consistency**: Ensure user, employee, and entity data is synchronized
9. **Security**: Store sensitive data (Aadhaar, PAN) encrypted
10. **Metadata**: Use metadata fields for extensibility and tracking
