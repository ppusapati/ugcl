package models

import (
	"time"

	"github.com/google/uuid"
)

// Domain models for tenant service

type Tenant struct {
	ID                    uuid.UUID  `json:"id"`
	Name                  string     `json:"name"`
	DisplayName           string     `json:"display_name"`
	Region                string     `json:"region"`
	Logo                  *string    `json:"logo,omitempty"`
	TenantDB              int32      `json:"tenant_db"` // 0=SHARED, 1=SEPERATEDB, 2=SEPERATESCHEMA
	IsActive              bool       `json:"is_active"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
	CreatedBy             *uuid.UUID `json:"created_by,omitempty"`
	SubscriptionPlan      *string    `json:"subscription_plan,omitempty"`
	SubscriptionExpiresAt *time.Time `json:"subscription_expires_at,omitempty"`
	MaxUsers              int32      `json:"max_users"`
	MaxStorageGB          int32      `json:"max_storage_gb"`
}

type TenantConnectionString struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	IsEncrypted bool      `json:"is_encrypted"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TenantFeature struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	ValueType   string    `json:"value_type"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TenantMetadata struct {
	ID            uuid.UUID              `json:"id"`
	TenantID      uuid.UUID              `json:"tenant_id"`
	Data          map[string]interface{} `json:"data"`
	SchemaVersion int32                  `json:"schema_version"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

type TenantDomain struct {
	ID                uuid.UUID  `json:"id"`
	TenantID          uuid.UUID  `json:"tenant_id"`
	Domain            string     `json:"domain"`
	IsPrimary         bool       `json:"is_primary"`
	IsVerified        bool       `json:"is_verified"`
	SSLCertificate    *string    `json:"ssl_certificate,omitempty"`
	SSLPrivateKey     *string    `json:"ssl_private_key,omitempty"`
	VerificationToken *string    `json:"verification_token,omitempty"`
	VerifiedAt        *time.Time `json:"verified_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type TenantAdminUser struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	UserID    string    `json:"user_id"`
	Username  *string   `json:"username,omitempty"`
	Email     *string   `json:"email,omitempty"`
	IsPrimary bool      `json:"is_primary"`
	CreatedAt time.Time `json:"created_at"`
}

type TenantDatabaseSchema struct {
	ID               uuid.UUID  `json:"id"`
	TenantID         uuid.UUID  `json:"tenant_id"`
	SchemaName       string     `json:"schema_name"`
	DatabaseName     string     `json:"database_name"`
	ConnectionString string     `json:"connection_string"`
	IsActive         bool       `json:"is_active"`
	CreatedAt        time.Time  `json:"created_at"`
	MigrationVersion *string    `json:"migration_version,omitempty"`
	LastMigrationAt  *time.Time `json:"last_migration_at,omitempty"`
}

type TenantUsageMetric struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	MetricName  string    `json:"metric_name"`
	MetricValue int64     `json:"metric_value"`
	MetricUnit  *string   `json:"metric_unit,omitempty"`
	RecordedAt  time.Time `json:"recorded_at"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
}

type TenantBilling struct {
	ID              uuid.UUID              `json:"id"`
	TenantID        uuid.UUID              `json:"tenant_id"`
	BillingEmail    string                 `json:"billing_email"`
	PaymentMethodID *string                `json:"payment_method_id,omitempty"`
	SubscriptionID  *string                `json:"subscription_id,omitempty"`
	BillingAddress  map[string]interface{} `json:"billing_address,omitempty"`
	TaxID           *string                `json:"tax_id,omitempty"`
	BillingCurrency string                 `json:"billing_currency"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

type TenantAuditEntry struct {
	ID        uuid.UUID              `json:"id"`
	TenantID  uuid.UUID              `json:"tenant_id"`
	Action    string                 `json:"action"`
	ActorID   *string                `json:"actor_id,omitempty"`
	ActorType string                 `json:"actor_type"`
	Changes   map[string]interface{} `json:"changes,omitempty"`
	IPAddress *string                `json:"ip_address,omitempty"`
	UserAgent *string                `json:"user_agent,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// Aggregate and summary models

type TenantSummary struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	DisplayName  string    `json:"display_name"`
	Region       string    `json:"region"`
	TenantDB     int32     `json:"tenant_db"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	AdminCount   int64     `json:"admin_count"`
	DomainCount  int64     `json:"domain_count"`
	FeatureCount int64     `json:"feature_count"`
}

type TenantStats struct {
	Tenant          *Tenant `json:"tenant"`
	VerifiedDomains int64   `json:"verified_domains"`
	FeatureCount    int64   `json:"feature_count"`
	RecentActivity  int64   `json:"recent_activity"`
}

type TenantByDomain struct {
	Domain      string    `json:"domain"`
	TenantID    uuid.UUID `json:"tenant_id"`
	TenantName  string    `json:"tenant_name"`
	DisplayName string    `json:"display_name"`
	IsVerified  bool      `json:"is_verified"`
}

type TenantWithBilling struct {
	TenantID    uuid.UUID      `json:"tenant_id"`
	TenantName  string         `json:"tenant_name"`
	DisplayName string         `json:"display_name"`
	Billing     *TenantBilling `json:"billing"`
}

type TenantDatabaseSchemaWithTenant struct {
	Schema      *TenantDatabaseSchema `json:"schema"`
	TenantName  string                `json:"tenant_name"`
	DisplayName string                `json:"display_name"`
}

type TenantAuditWithTenant struct {
	AuditEntry  *TenantAuditEntry `json:"audit_entry"`
	TenantName  string            `json:"tenant_name"`
	DisplayName string            `json:"display_name"`
}

// Analytics and reporting models

type TenantRegionStats struct {
	Region              string `json:"region"`
	TenantCount         int64  `json:"tenant_count"`
	ActiveCount         int64  `json:"active_count"`
	SeparateDBCount     int64  `json:"separate_db_count"`
	SeparateSchemaCount int64  `json:"separate_schema_count"`
}

type TenantSubscriptionStats struct {
	SubscriptionPlan string `json:"subscription_plan"`
	TenantCount      int64  `json:"tenant_count"`
	ActiveCount      int64  `json:"active_count"`
	ExpiredCount     int64  `json:"expired_count"`
}

type TenantDashboardStats struct {
	TotalActiveTenants    int64 `json:"total_active_tenants"`
	SharedDBTenants       int64 `json:"shared_db_tenants"`
	SeparateDBTenants     int64 `json:"separate_db_tenants"`
	SeparateSchemaTenants int64 `json:"separate_schema_tenants"`
	VerifiedDomains       int64 `json:"verified_domains"`
	TenantsWithFeatures   int64 `json:"tenants_with_features"`
}

type TenantGrowthStats struct {
	Period         time.Time `json:"period"`
	NewTenants     int64     `json:"new_tenants"`
	SharedDB       int64     `json:"shared_db"`
	SeparateDB     int64     `json:"separate_db"`
	SeparateSchema int64     `json:"separate_schema"`
}

type TenantFeatureUsage struct {
	FeatureName  string   `json:"feature_name"`
	TenantCount  int64    `json:"tenant_count"`
	EnabledCount int64    `json:"enabled_count"`
	UniqueValues []string `json:"unique_values"`
}

type TenantsBySubscriptionPlan struct {
	SubscriptionPlan   string  `json:"subscription_plan"`
	TenantCount        int64   `json:"tenant_count"`
	ActiveCount        int64   `json:"active_count"`
	ValidSubscriptions int64   `json:"valid_subscriptions"`
	AvgMaxUsers        float64 `json:"avg_max_users"`
	AvgMaxStorage      float64 `json:"avg_max_storage"`
}

type TenantActivitySummary struct {
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	DisplayName    string     `json:"display_name"`
	CreatedAt      time.Time  `json:"created_at"`
	RecentActivity int64      `json:"recent_activity"`
	LastActivity   *time.Time `json:"last_activity,omitempty"`
}

type TenantUsageMetricsSummary struct {
	MetricName  string  `json:"metric_name"`
	MetricUnit  *string `json:"metric_unit,omitempty"`
	TotalValue  int64   `json:"total_value"`
	AvgValue    float64 `json:"avg_value"`
	MaxValue    int64   `json:"max_value"`
	MinValue    int64   `json:"min_value"`
	RecordCount int64   `json:"record_count"`
}

// Utility models
type OrphanedData struct {
	TableName string `json:"table_name"`
	TenantID  string `json:"tenant_id"`
}

type TenantDataSize struct {
	ID                     uuid.UUID `json:"id"`
	Name                   string    `json:"name"`
	FeaturesCount          int64     `json:"features_count"`
	ConnectionStringsCount int64     `json:"connection_strings_count"`
	DomainsCount           int64     `json:"domains_count"`
	UsageMetricsCount      int64     `json:"usage_metrics_count"`
	AuditEntriesCount      int64     `json:"audit_entries_count"`
}
