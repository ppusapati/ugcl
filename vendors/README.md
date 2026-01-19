# Vendors Management Module

## 1. Module Overview

The Vendors module manages supplier and vendor companies that provide materials, equipment, and services to UGCL. It handles vendor profiles, payment terms, banking information, ratings, contracts, and blacklisting capabilities.

**Purpose:** Centralized vendor relationship management with focus on procurement, payment processing, performance tracking, and compliance management.

**Key Features:**
- Vendor company profile management
- Category-based vendor classification
- Payment terms and credit period tracking
- Bank account management for payments
- Vendor rating system (1-5 stars)
- Blacklist management
- Contract and purchase order tracking
- Statutory compliance (GST, PAN)
- Contact person management via user accounts
- Vendor performance analytics

## 2. Architecture

### Module Structure
```
vendors/
├── proto/
│   └── vendor.proto         # Service and message definitions
├── db/
│   ├── schema/
│   │   └── vendors.sql     # Database schema
│   └── generated/          # SQLC generated code
├── repository/             # Data access layer
│   └── vendor_repository.go
├── services/               # Business logic layer
│   └── vendor_service.go
├── handlers/               # gRPC handlers
│   └── vendor_handler.go
├── mappers/                # DTO mappers
│   └── vendor_mapper.go
└── module.go               # FX module definition
```

### Database Schema

**Tables:**
- `vendors` - Vendor company information and payment details

**Key Relationships:**
- Vendor → User (N:1 via person_id for contact person)
- Vendor → Contracts (N:N via contracts array)
- Vendor → Purchase Orders (N:N via purchase_orders array)

### Key Dependencies
- `identity/user` - User account for contact persons
- `identity/entity` - Entity abstraction
- `procurement` - Purchase orders (future)
- Connect-Go RPC framework
- PostgreSQL with array and JSONB support

## 3. Quick Start

### Create a Vendor

```go
import (
    vendorv2 "p9e.in/ugcl/vendors/api/v2/vendor"
    userv2 "p9e.in/ugcl/identity/user/api/v2/user"
    "connectrpc.com/connect"
)

client := vendorv2.NewVendorServiceClient(httpClient, baseURL)

// Create vendor with contact person
resp, err := client.CreateVendor(ctx, connect.NewRequest(&vendorv2.CreateVendorRequest{
    Vendor: &vendorv2.Vendor{
        CompanyName:       "Steel Suppliers India Ltd",
        CompanyType:       "Private Limited",
        Gst:               "27AABCS9603R1ZM",
        Pan:               "AABCS9603R",
        VendorCategory:    "RAW_MATERIAL",
        PaymentTerms:      "Net 30",
        CreditPeriodDays:  30,
        BankName:          "State Bank of India",
        AccountNumber:     "12345678901234",
        Ifsc:              "SBIN0001234",
        Rating:            4,
        IsBlacklisted:     false,
        Status:            "ACTIVE",
        Metadata: &structpb.Struct{
            Fields: map[string]*structpb.Value{
                "specialization": structpb.NewStringValue("Steel, Iron, Metal"),
                "capacity":       structpb.NewStringValue("5000 tons/month"),
            },
        },
    },
    User: &userv2.User{
        Username:       wrapperspb.String("contact.steel"),
        Email:          wrapperspb.String("sales@steelsuppliers.com"),
        Fullname:       wrapperspb.String("Rajesh Sharma"),
        Phone:          wrapperspb.String("+91-9876543210"),
        Password:       "VendorPassword123!",
        ConfirmPassword: "VendorPassword123!",
    },
}))
```

### Common Operations

**Get vendor:**
```go
vendor, err := client.GetVendor(ctx, connect.NewRequest(&vendorv2.VendorIdentifier{
    Id: "vendor-uuid",
}))
```

**List vendors by category:**
```go
vendors, err := client.GetVendorsByCategory(ctx, connect.NewRequest(&vendorv2.CategoryIdentifier{
    Category: "RAW_MATERIAL",
}))
```

**Update vendor rating:**
```go
updated, err := client.UpdateVendor(ctx, connect.NewRequest(&vendorv2.UpdateVendorRequest{
    Vendor: &vendorv2.Vendor{
        Id:     vendorID,
        Rating: 5,
    },
    UpdateMask: &fieldmaskpb.FieldMask{
        Paths: []string{"rating"},
    },
}))
```

**Blacklist vendor:**
```go
blacklisted, err := client.UpdateVendor(ctx, connect.NewRequest(&vendorv2.UpdateVendorRequest{
    Vendor: &vendorv2.Vendor{
        Id:            vendorID,
        IsBlacklisted: true,
        Status:        "BLACKLISTED",
    },
    UpdateMask: &fieldmaskpb.FieldMask{
        Paths: []string{"is_blacklisted", "status"},
    },
}))
```

## 4. API Reference

### Vendor Operations
- `CreateVendor` - Create new vendor with contact person
- `UpdateVendor` - Update vendor details with field mask
- `GetVendor` - Get vendor by ID
- `ListVendors` - List vendors with filtering and pagination
- `DeleteVendor` - Soft delete vendor record
- `GetVendorByUserId` - Find vendor by contact person user ID
- `GetVendorsByCategory` - List vendors in specific category

## 5. Database Schema

### Table: vendors

```sql
-- Identity fields
- id (BIGSERIAL, PK)
- uuid (UUID, UNIQUE, generated)
- company_name (VARCHAR(255))
- company_type (VARCHAR(100)) -- Private Limited, Partnership, Proprietorship, etc.

-- Statutory information
- gst (VARCHAR(50)) -- GST number
- pan (VARCHAR(50), UNIQUE) -- PAN of company
- vendor_category (VARCHAR(100)) -- RAW_MATERIAL, EQUIPMENT, SERVICES, CONSUMABLES

-- Contact information
- person_id (BIGINT, FK → users.id) -- Contact person user account

-- Payment details
- payment_terms (TEXT) -- Net 30, Net 60, Advance, etc.
- credit_period_days (INTEGER) -- Credit period in days
- bank_name (VARCHAR(255))
- account_number (VARCHAR(100))
- ifsc (VARCHAR(20))

-- Performance tracking
- rating (INTEGER) -- 1-5 stars, CHECK constraint
- is_blacklisted (BOOLEAN) -- Blacklist flag

-- Business relationships
- contracts (TEXT[]) -- Array of contract IDs
- purchase_orders (TEXT[]) -- Array of PO IDs

-- Status and metadata
- status (VARCHAR(50)) -- ACTIVE, INACTIVE, SUSPENDED, BLACKLISTED
- metadata (JSONB) -- Additional information

-- Audit fields
- created_at, updated_at (TIMESTAMP)
- deleted_at (TIMESTAMP, soft delete)
```

### Constraints
- `chk_rating` - rating >= 0 AND rating <= 5
- `chk_credit_period` - credit_period_days >= 0
- PAN must be unique across all vendors

### Indexes
- `idx_vendors_person_id` - Contact person lookup
- `idx_vendors_status` - Filter by status
- `idx_vendors_company` - Company name search
- `idx_vendors_category` - Category filtering
- `idx_vendors_pan` - PAN lookup
- `idx_vendors_gst` - GST lookup
- `idx_vendors_rating` - Rating-based queries
- `idx_vendors_blacklisted` - Blacklist filtering
- `idx_vendors_contracts` - GIN index for contract associations
- `idx_vendors_purchase_orders` - GIN index for PO tracking

## 6. Configuration

### Environment Variables
```env
# Database connection managed by packages/database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ugcl
```

### Module-Specific Settings
- Default credit period (configurable per category)
- Rating calculation parameters
- Blacklist notification settings
- Payment terms templates

## 7. Examples

### Complete Vendor Lifecycle

```go
// 1. Create new vendor
vendor, err := client.CreateVendor(ctx, connect.NewRequest(&vendorv2.CreateVendorRequest{
    Vendor: &vendorv2.Vendor{
        CompanyName:      "Equipment Rentals Co",
        CompanyType:      "Partnership",
        Gst:              "27ABCDE1234F1Z5",
        Pan:              "ABCDE1234F",
        VendorCategory:   "EQUIPMENT",
        PaymentTerms:     "Net 60",
        CreditPeriodDays: 60,
        BankName:         "HDFC Bank",
        AccountNumber:    "98765432109876",
        Ifsc:             "HDFC0001234",
        Rating:           3,
        Status:           "ACTIVE",
    },
    User: &userv2.User{
        Username:       wrapperspb.String("contact.equipment"),
        Email:          wrapperspb.String("sales@equipmentrentals.com"),
        Fullname:       wrapperspb.String("Amit Verma"),
        Phone:          wrapperspb.String("+91-9123456789"),
        Password:       "Password123!",
        ConfirmPassword: "Password123!",
    },
}))

vendorID := vendor.Msg.Id

// 2. Add contracts and purchase orders
assigned, err := client.UpdateVendor(ctx, connect.NewRequest(&vendorv2.UpdateVendorRequest{
    Vendor: &vendorv2.Vendor{
        Id:             vendorID,
        Contracts:      []string{"contract-uuid-1", "contract-uuid-2"},
        PurchaseOrders: []string{"po-uuid-1", "po-uuid-2"},
    },
    UpdateMask: &fieldmaskpb.FieldMask{
        Paths: []string{"contracts", "purchase_orders"},
    },
}))

// 3. Update rating based on performance
ratingUpdate, err := client.UpdateVendor(ctx, connect.NewRequest(&vendorv2.UpdateVendorRequest{
    Vendor: &vendorv2.Vendor{
        Id:     vendorID,
        Rating: 5, // Excellent performance
    },
    UpdateMask: &fieldmaskpb.FieldMask{
        Paths: []string{"rating"},
    },
}))

// 4. Blacklist vendor due to quality issues
blacklisted, err := client.UpdateVendor(ctx, connect.NewRequest(&vendorv2.UpdateVendorRequest{
    Vendor: &vendorv2.Vendor{
        Id:            vendorID,
        IsBlacklisted: true,
        Status:        "BLACKLISTED",
        Metadata: &structpb.Struct{
            Fields: map[string]*structpb.Value{
                "blacklist_reason": structpb.NewStringValue("Quality issues - multiple complaints"),
                "blacklist_date":   structpb.NewStringValue(time.Now().Format("2006-01-02")),
            },
        },
    },
    UpdateMask: &fieldmaskpb.FieldMask{
        Paths: []string{"is_blacklisted", "status", "metadata"},
    },
}))
```

### Vendor Performance Analysis

```go
// Get top-rated vendors by category
vendors, err := client.ListVendors(ctx, connect.NewRequest(&vendorv2.ListVendorsRequest{
    Filter: &vendorv2.VendorFilter{
        VendorCategory: "RAW_MATERIAL",
        MinRating:      4,
        Status:         "ACTIVE",
        IsBlacklisted:  false,
    },
    Sort: []string{"-rating", "company_name"},
}))

fmt.Println("Top-rated Raw Material Vendors:")
for _, v := range vendors.Msg.Vendors {
    fmt.Printf("- %s (Rating: %d/5, Credit: %d days)\n",
        v.CompanyName, v.Rating, v.CreditPeriodDays)
}
```

### Payment Terms Reporting

```go
// Get vendors with extended credit periods
vendors, err := client.ListVendors(ctx, connect.NewRequest(&vendorv2.ListVendorsRequest{
    Filter: &vendorv2.VendorFilter{
        Status: "ACTIVE",
    },
}))

// Application-level filtering for credit > 30 days
var extendedCredit []*vendorv2.Vendor
for _, v := range vendors.Msg.Vendors {
    if v.CreditPeriodDays > 30 {
        extendedCredit = append(extendedCredit, v)
    }
}

fmt.Printf("Vendors with >30 days credit: %d\n", len(extendedCredit))
```

### Blacklist Management

```go
// Get all blacklisted vendors
blacklisted, err := client.ListVendors(ctx, connect.NewRequest(&vendorv2.ListVendorsRequest{
    Filter: &vendorv2.VendorFilter{
        IsBlacklisted: true,
    },
}))

fmt.Println("Blacklisted Vendors:")
for _, v := range blacklisted.Msg.Vendors {
    reason := "N/A"
    if v.Metadata != nil && v.Metadata.Fields["blacklist_reason"] != nil {
        reason = v.Metadata.Fields["blacklist_reason"].GetStringValue()
    }
    fmt.Printf("- %s: %s\n", v.CompanyName, reason)
}
```

## 8. Integration

### With Other Modules

**Identity/User Module:**
- Contact person has user account for vendor portal access
- User.id → Vendor.person_id relationship
- Contact person can view orders and invoices

**Identity/Entity Module:**
- Vendor record creates corresponding entity
- Entity.reference_id → Vendor.uuid
- Entity type: VENDOR

**Procurement Module (Future):**
- Vendors receive purchase orders
- purchase_orders array tracks PO associations
- Vendor selection based on rating and category

**Finance/Accounts Module (Future):**
- Vendor payments use bank account info
- Credit period affects payment scheduling
- Invoice reconciliation with POs

**DMS Module:**
- Vendor documents (contracts, licenses, certifications)
- Document ownership via entity
- Compliance document management

**Inventory Module (Future):**
- Materials received from vendors
- Quality checks update vendor ratings
- Vendor performance metrics

### Event Publishing
```go
// Vendor lifecycle events
type VendorEvent struct {
    Type       string    // created, rated, blacklisted, activated, suspended
    VendorID   string
    CompanyName string
    Category   string
    Timestamp  time.Time
    Changes    map[string]interface{}
}
```

## 9. Development

### Modifying the Module

**Add New Field:**
1. Update `vendors.sql` schema
2. Update `vendor.proto` message
3. Run migration: `make migrate-up`
4. Regenerate SQLC: `sqlc generate -f vendors/db/sqlc.yaml`
5. Regenerate proto: `buf generate`
6. Update mappers and services

**Add Vendor Category:**
```sql
-- Categories stored as VARCHAR, add validation
-- Common categories: RAW_MATERIAL, EQUIPMENT, SERVICES, CONSUMABLES, PACKAGING, LOGISTICS
ALTER TABLE vendors ADD CONSTRAINT chk_vendor_category
    CHECK (vendor_category IN (
        'RAW_MATERIAL', 'EQUIPMENT', 'SERVICES', 'CONSUMABLES',
        'PACKAGING', 'LOGISTICS', 'UTILITIES', 'OTHER'
    ));
```

### Generate Code
```bash
# Generate proto stubs
buf generate

# Generate SQLC queries
sqlc generate -f vendors/db/sqlc.yaml

# Run migrations
make migrate-up
```

### Testing Guidelines

**Unit Tests:**
```go
func TestCreateVendor(t *testing.T) {
    mockRepo := &MockVendorRepository{}
    mockUserService := &MockUserService{}
    svc := services.NewVendorService(mockRepo, mockUserService)

    vendor, err := svc.CreateVendor(ctx, req)
    assert.NoError(t, err)
    assert.Equal(t, "Steel Suppliers", vendor.CompanyName)
    assert.Equal(t, 4, vendor.Rating)
}
```

**Integration Tests:**
- Test vendor creation with user
- Test rating updates
- Test blacklist functionality
- Test soft delete behavior

## 10. Troubleshooting

### Common Issues

**Issue: Duplicate PAN**
- **Cause:** PAN already registered with another vendor
- **Solution:** Check existing vendors before creation

**Issue: Invalid rating value**
- **Cause:** Rating outside 0-5 range
- **Solution:** Database constraint will reject, validate before save

**Issue: Negative credit period**
- **Cause:** credit_period_days < 0
- **Solution:** Validate input, constraint will catch negative values

**Issue: User creation fails for contact person**
- **Cause:** Email/username already exists
- **Solution:** Validate user uniqueness before vendor creation

**Issue: Cannot delete vendor**
- **Cause:** Vendor has active purchase orders or pending payments
- **Solution:** Complete pending transactions, or use soft delete

### Performance Tips

1. Index frequently queried fields (company_name, category, rating)
2. Use GIN indexes for array fields (contracts, purchase_orders)
3. Cache active vendors list by category
4. Batch vendor imports with transactions
5. Use partial indexes on deleted_at for active queries

### Debugging

**Check vendor existence:**
```sql
SELECT uuid, company_name, vendor_category, rating, status
FROM vendors
WHERE pan = 'AABCS9603R' AND deleted_at IS NULL;
```

**Find top-rated vendors:**
```sql
SELECT company_name, vendor_category, rating, credit_period_days
FROM vendors
WHERE rating >= 4
  AND is_blacklisted = FALSE
  AND status = 'ACTIVE'
  AND deleted_at IS NULL
ORDER BY rating DESC, company_name;
```

**Verify payment terms:**
```sql
SELECT company_name, payment_terms, credit_period_days,
       bank_name, account_number
FROM vendors
WHERE status = 'ACTIVE'
  AND deleted_at IS NULL
ORDER BY credit_period_days DESC;
```

**Blacklist audit:**
```sql
SELECT company_name, is_blacklisted, status,
       metadata->>'blacklist_reason' as reason,
       metadata->>'blacklist_date' as blacklisted_on
FROM vendors
WHERE is_blacklisted = TRUE
  AND deleted_at IS NULL;
```

### Business Rules

1. **Rating System:** 1 (Poor) to 5 (Excellent)
2. **Blacklisting:** Automatically sets status to BLACKLISTED
3. **Credit Period:** Must be >= 0, typically 0, 15, 30, 45, 60, 90 days
4. **Payment Terms:** Common values: Advance, Net 15, Net 30, Net 60, Net 90
5. **Status Flow:** ACTIVE → SUSPENDED → ACTIVE or ACTIVE → BLACKLISTED

### Validation Rules

**GST Number:**
- Format: 15 characters (2 state + 10 PAN + 1 entity + 1 Z + 1 checksum)
- Example: 27AABCS9603R1ZM

**PAN Number:**
- Format: 10 characters (AAA + C + A + 9999 + A)
- Example: AABCS9603R

**IFSC Code:**
- Format: 11 characters (ABCD0123456)
- First 4: Bank code
- 5th: Always 0
- Last 6: Branch code

**Rating:**
- Integer between 0 and 5 (inclusive)
- 0 = Not rated, 1 = Poor, 5 = Excellent

## Additional Resources

- [GST Portal](https://www.gst.gov.in/)
- [PAN Verification](https://www.incometax.gov.in/iec/foportal/)
- [IFSC Code Search](https://www.rbi.org.in/)
- [Vendor Management Best Practices](https://www.cips.org/supply-management/analysis/2021/march/vendor-management-best-practices/)
- [Payment Terms Guide](https://www.investopedia.com/terms/n/net-30.asp)
