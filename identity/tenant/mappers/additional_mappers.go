package mappers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	sqlcgen "p9e.in/ugcl/identity/tenant/db/generated"
	"p9e.in/ugcl/identity/tenant/models"
)

// Additional mapping functions for all tenant domain models

// TenantAdminUser mappings
func (m *TenantMapper) FromSQLCTenantAdminUser(sqlcAdminUser sqlcgen.TenantAdminUser) *models.TenantAdminUser {
	adminUser := &models.TenantAdminUser{
		ID:        sqlcAdminUser.ID,
		TenantID:  sqlcAdminUser.TenantID,
		UserID:    sqlcAdminUser.UserID,
		IsPrimary: *sqlcAdminUser.IsPrimary,
		CreatedAt: sqlcAdminUser.CreatedAt,
	}

	if sqlcAdminUser.Username != nil {
		adminUser.Username = sqlcAdminUser.Username
	}
	if sqlcAdminUser.Email != nil {
		adminUser.Email = sqlcAdminUser.Email
	}

	return adminUser
}

// TenantBilling mappings
func (m *TenantMapper) FromSQLCTenantBilling(sqlcBilling sqlcgen.TenantBilling) (*models.TenantBilling, error) {
	var billingCurrency string
	if sqlcBilling.BillingCurrency != nil {
		billingCurrency = *sqlcBilling.BillingCurrency
	}

	billing := &models.TenantBilling{
		ID:              sqlcBilling.ID,
		TenantID:        sqlcBilling.TenantID,
		BillingEmail:    sqlcBilling.BillingEmail,
		BillingCurrency: billingCurrency,
		CreatedAt:       sqlcBilling.CreatedAt,
		UpdatedAt:       sqlcBilling.UpdatedAt,
	}

	if sqlcBilling.PaymentMethodID != nil {
		billing.PaymentMethodID = sqlcBilling.PaymentMethodID
	}
	if sqlcBilling.SubscriptionID != nil {
		billing.SubscriptionID = sqlcBilling.SubscriptionID
	}
	if sqlcBilling.TaxID != nil {
		billing.TaxID = sqlcBilling.TaxID
	}

	// Parse JSONB billing address
	if len(sqlcBilling.BillingAddress) > 0 {
		var address map[string]interface{}
		if err := json.Unmarshal(sqlcBilling.BillingAddress, &address); err != nil {
			return nil, fmt.Errorf("failed to unmarshal billing address: %w", err)
		}
		billing.BillingAddress = address
	}

	return billing, nil
}

// TenantUsageMetric mappings
func (m *TenantMapper) FromSQLCTenantUsageMetric(sqlcMetric sqlcgen.TenantUsageMetric) *models.TenantUsageMetric {
	metric := &models.TenantUsageMetric{
		ID:          sqlcMetric.ID,
		TenantID:    sqlcMetric.TenantID,
		MetricName:  sqlcMetric.MetricName,
		MetricValue: sqlcMetric.MetricValue,
		RecordedAt:  sqlcMetric.RecordedAt.Time,
		PeriodStart: sqlcMetric.PeriodStart.Time,
		PeriodEnd:   sqlcMetric.PeriodEnd.Time,
	}

	if sqlcMetric.MetricUnit != nil {
		metric.MetricUnit = sqlcMetric.MetricUnit
	}

	return metric
}

// TenantAuditEntry mappings
func (m *TenantMapper) FromSQLCTenantAuditEntry(sqlcEntry sqlcgen.TenantAuditLog) (*models.TenantAuditEntry, error) {
	entry := &models.TenantAuditEntry{
		ID:        sqlcEntry.ID,
		TenantID:  sqlcEntry.TenantID,
		Action:    sqlcEntry.Action,
		ActorType: *sqlcEntry.ActorType,
		CreatedAt: sqlcEntry.CreatedAt,
	}

	if sqlcEntry.ActorID != nil {
		entry.ActorID = sqlcEntry.ActorID
	}
	if sqlcEntry.IpAddress != nil {
		ipStr := sqlcEntry.IpAddress.String()
		entry.IPAddress = &ipStr
	}
	if sqlcEntry.UserAgent != nil {
		entry.UserAgent = sqlcEntry.UserAgent
	}

	// Parse JSONB changes
	if len(sqlcEntry.Changes) > 0 {
		var changes map[string]interface{}
		if err := json.Unmarshal(sqlcEntry.Changes, &changes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal audit entry changes: %w", err)
		}
		entry.Changes = changes
	}

	return entry, nil
}

// TenantDatabaseSchema mappings
func (m *TenantMapper) FromSQLCTenantDatabaseSchema(sqlcSchema sqlcgen.TenantDatabaseSchema) *models.TenantDatabaseSchema {
	schema := &models.TenantDatabaseSchema{
		ID:               sqlcSchema.ID,
		TenantID:         sqlcSchema.TenantID,
		SchemaName:       sqlcSchema.SchemaName,
		DatabaseName:     sqlcSchema.DatabaseName,
		ConnectionString: sqlcSchema.ConnectionString,
		IsActive:         *sqlcSchema.IsActive,
		CreatedAt:        sqlcSchema.CreatedAt,
	}

	if sqlcSchema.MigrationVersion != nil {
		schema.MigrationVersion = sqlcSchema.MigrationVersion
	}
	if sqlcSchema.LastMigrationAt.Valid {
		schema.LastMigrationAt = &sqlcSchema.LastMigrationAt.Time
	}

	return schema
}

// Reporting and analytics mappings

func (m *TenantMapper) FromSQLCTenantDashboardStats(sqlcStats sqlcgen.GetTenantDashboardStatsRow) *models.TenantDashboardStats {
	return &models.TenantDashboardStats{
		TotalActiveTenants:    sqlcStats.TotalActiveTenants,
		SharedDBTenants:       sqlcStats.SharedDbTenants,
		SeparateDBTenants:     sqlcStats.SeparateDbTenants,
		SeparateSchemaTenants: sqlcStats.SeparateSchemaTenants,
		VerifiedDomains:       sqlcStats.VerifiedDomains,
		TenantsWithFeatures:   sqlcStats.TenantsWithFeatures,
	}
}

func (m *TenantMapper) FromSQLCTenantGrowthStats(sqlcStats sqlcgen.GetTenantGrowthStatsRow) *models.TenantGrowthStats {
	// Convert interval to a base time for period representation
	period := time.Now() // This should be adjusted based on your interval logic

	return &models.TenantGrowthStats{
		Period:         period,
		NewTenants:     sqlcStats.NewTenants,
		SharedDB:       sqlcStats.SharedDb,
		SeparateDB:     sqlcStats.SeparateDb,
		SeparateSchema: sqlcStats.SeparateSchema,
	}
}

func (m *TenantMapper) FromSQLCTenantFeatureUsage(sqlcStats sqlcgen.GetTenantFeatureUsageRow) *models.TenantFeatureUsage {
	var uniqueValues []string
	if sqlcStats.UniqueValues != nil {
		if vals, ok := sqlcStats.UniqueValues.([]interface{}); ok {
			for _, v := range vals {
				if str, ok := v.(string); ok {
					uniqueValues = append(uniqueValues, str)
				}
			}
		}
	}

	return &models.TenantFeatureUsage{
		FeatureName:  sqlcStats.FeatureName,
		TenantCount:  sqlcStats.TenantCount,
		EnabledCount: sqlcStats.EnabledCount,
		UniqueValues: uniqueValues,
	}
}

func (m *TenantMapper) FromSQLCTenantsBySubscriptionPlan(sqlcStats sqlcgen.GetTenantsBySubscriptionPlanRow) *models.TenantsBySubscriptionPlan {
	var subscriptionPlan string
	if sqlcStats.SubscriptionPlan != nil {
		subscriptionPlan = *sqlcStats.SubscriptionPlan
	}

	return &models.TenantsBySubscriptionPlan{
		SubscriptionPlan:   subscriptionPlan,
		TenantCount:        sqlcStats.TenantCount,
		ActiveCount:        sqlcStats.ActiveCount,
		ValidSubscriptions: sqlcStats.ValidSubscriptions,
		AvgMaxUsers:        sqlcStats.AvgMaxUsers,
		AvgMaxStorage:      sqlcStats.AvgMaxStorage,
	}
}

func (m *TenantMapper) FromSQLCTenantActivitySummary(sqlcStats sqlcgen.GetTenantActivitySummaryRow) *models.TenantActivitySummary {
	summary := &models.TenantActivitySummary{
		ID:             sqlcStats.ID,
		Name:           sqlcStats.Name,
		DisplayName:    sqlcStats.DisplayName,
		CreatedAt:      sqlcStats.CreatedAt,
		RecentActivity: sqlcStats.RecentActivity,
	}

	if sqlcStats.LastActivity != nil {
		if lastActivity, ok := sqlcStats.LastActivity.(time.Time); ok {
			summary.LastActivity = &lastActivity
		}
	}

	return summary
}

func (m *TenantMapper) FromSQLCTenantUsageMetricsSummary(sqlcStats sqlcgen.GetTenantUsageMetricsSummaryRow) *models.TenantUsageMetricsSummary {
	var maxValue, minValue int64
	if sqlcStats.MaxValue != nil {
		if val, ok := sqlcStats.MaxValue.(int64); ok {
			maxValue = val
		}
	}
	if sqlcStats.MinValue != nil {
		if val, ok := sqlcStats.MinValue.(int64); ok {
			minValue = val
		}
	}

	summary := &models.TenantUsageMetricsSummary{
		MetricName:  sqlcStats.MetricName,
		TotalValue:  sqlcStats.TotalValue,
		AvgValue:    sqlcStats.AvgValue,
		MaxValue:    maxValue,
		MinValue:    minValue,
		RecordCount: sqlcStats.RecordCount,
	}

	if sqlcStats.MetricUnit != nil {
		summary.MetricUnit = sqlcStats.MetricUnit
	}

	return summary
}

// Utility mappings
func (m *TenantMapper) FromSQLCOrphanedData(sqlcData sqlcgen.GetOrphanedTenantDataRow) *models.OrphanedData {
	return &models.OrphanedData{
		TableName: sqlcData.TableName,
		TenantID:  uuid.UUID(sqlcData.TenantID).String(),
	}
}

func (m *TenantMapper) FromSQLCTenantDataSize(sqlcData sqlcgen.GetTenantDataSizesRow) *models.TenantDataSize {
	return &models.TenantDataSize{
		ID:                     sqlcData.ID,
		Name:                   sqlcData.Name,
		FeaturesCount:          sqlcData.FeaturesCount,
		ConnectionStringsCount: sqlcData.ConnectionStringsCount,
		DomainsCount:           sqlcData.DomainsCount,
		UsageMetricsCount:      sqlcData.UsageMetricsCount,
		AuditEntriesCount:      sqlcData.AuditEntriesCount,
	}
}
