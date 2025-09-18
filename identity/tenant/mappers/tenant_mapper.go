package mappers

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	tenantpb "p9e.in/ugcl/identity/tenant/api/v1/tenant"
	sqlcgen "p9e.in/ugcl/identity/tenant/db/generated"
	"p9e.in/ugcl/identity/tenant/models"
)

type TenantMapper struct{}

func NewTenantMapper() *TenantMapper {
	return &TenantMapper{}
}

// Tenant mapping functions

func (m *TenantMapper) FromSQLCTenant(sqlcTenant sqlcgen.Tenant) *models.Tenant {
	tenant := &models.Tenant{
		ID:           sqlcTenant.ID,
		Name:         sqlcTenant.Name,
		DisplayName:  sqlcTenant.DisplayName,
		Region:       sqlcTenant.Region,
		TenantDB:     sqlcTenant.TenantDb.(int32),
		IsActive:     *sqlcTenant.IsActive,
		CreatedAt:    sqlcTenant.CreatedAt,
		UpdatedAt:    sqlcTenant.UpdatedAt,
		MaxUsers:     *sqlcTenant.MaxUsers,
		MaxStorageGB: *sqlcTenant.MaxStorageGb,
	}

	if sqlcTenant.Logo != nil {
		tenant.Logo = sqlcTenant.Logo
	}
	if sqlcTenant.CreatedBy.Valid {
		createdBy := uuid.UUID(sqlcTenant.CreatedBy.Bytes)
		tenant.CreatedBy = &createdBy
	}
	if sqlcTenant.SubscriptionPlan != nil {
		tenant.SubscriptionPlan = sqlcTenant.SubscriptionPlan
	}
	if sqlcTenant.SubscriptionExpiresAt.Valid {
		tenant.SubscriptionExpiresAt = &sqlcTenant.SubscriptionExpiresAt.Time
	}

	return tenant
}

func (m *TenantMapper) ToCreateTenantParams(tenant *models.Tenant) sqlcgen.CreateTenantParams {
	params := sqlcgen.CreateTenantParams{
		Name:         tenant.Name,
		DisplayName:  tenant.DisplayName,
		Region:       tenant.Region,
		TenantDb:     tenant.TenantDB,
		MaxUsers:     &tenant.MaxUsers,
		MaxStorageGb: &tenant.MaxStorageGB,
	}

	if tenant.Logo != nil {
		params.Logo = tenant.Logo
	}
	if tenant.CreatedBy != nil {
		params.CreatedBy = pgtype.UUID{Bytes: *tenant.CreatedBy, Valid: true}
	}
	if tenant.SubscriptionPlan != nil {
		params.SubscriptionPlan = tenant.SubscriptionPlan
	}

	return params
}

func (m *TenantMapper) ToUpdateTenantParams(tenant *models.Tenant) sqlcgen.UpdateTenantParams {
	params := sqlcgen.UpdateTenantParams{
		ID:           tenant.ID,
		DisplayName:  tenant.DisplayName,
		MaxUsers:     &tenant.MaxUsers,
		MaxStorageGb: &tenant.MaxStorageGB,
	}

	if tenant.Logo != nil {
		params.Logo = tenant.Logo
	}
	if tenant.SubscriptionPlan != nil {
		params.SubscriptionPlan = tenant.SubscriptionPlan
	}

	return params
}

func (m *TenantMapper) ToProtobufTenant(tenant *models.Tenant) *tenantpb.Tenant {
	pbTenant := &tenantpb.Tenant{
		Id:          tenant.ID.String(),
		Name:        tenant.Name,
		DisplayName: tenant.DisplayName,
		Region:      tenant.Region,
		TenantDb:    tenantpb.TenantDB(tenant.TenantDB),
	}

	if tenant.Logo != nil {
		pbTenant.Logo = *tenant.Logo
	}

	return pbTenant
}

func (m *TenantMapper) FromProtobufCreateTenantRequest(req *tenantpb.CreateTenantRequest) *models.Tenant {
	tenant := &models.Tenant{
		Name:         req.Name,
		DisplayName:  req.DisplayName,
		Region:       req.Region,
		Logo:         &req.Logo,
		TenantDB:     0,   // Default to shared
		MaxUsers:     100, // Default limits
		MaxStorageGB: 10,
	}

	if req.SeparateDb {
		tenant.TenantDB = 1 // SEPERATEDB
	}

	return tenant
}

// TenantFeature mapping functions

func (m *TenantMapper) FromSQLCTenantFeature(sqlcFeature sqlcgen.TenantFeature) *models.TenantFeature {
	var valueType string
	if sqlcFeature.ValueType != nil {
		valueType = *sqlcFeature.ValueType
	}

	feature := &models.TenantFeature{
		ID:        sqlcFeature.ID,
		TenantID:  sqlcFeature.TenantID,
		Key:       sqlcFeature.Key,
		Value:     sqlcFeature.Value,
		ValueType: valueType,
		CreatedAt: sqlcFeature.CreatedAt,
		UpdatedAt: sqlcFeature.UpdatedAt,
	}

	if sqlcFeature.Description != nil {
		feature.Description = sqlcFeature.Description
	}

	return feature
}

func (m *TenantMapper) ToCreateTenantFeatureParams(feature *models.TenantFeature) sqlcgen.CreateTenantFeatureParams {
	params := sqlcgen.CreateTenantFeatureParams{
		TenantID:  feature.TenantID,
		Key:       feature.Key,
		Value:     feature.Value,
		ValueType: &feature.ValueType,
	}

	if feature.Description != nil {
		params.Description = feature.Description
	}

	return params
}

func (m *TenantMapper) ToUpdateTenantFeatureParams(feature *models.TenantFeature) sqlcgen.UpdateTenantFeatureParams {
	params := sqlcgen.UpdateTenantFeatureParams{
		TenantID:  feature.TenantID,
		Key:       feature.Key,
		Value:     feature.Value,
		ValueType: &feature.ValueType,
	}

	if feature.Description != nil {
		params.Description = feature.Description
	}

	return params
}

func (m *TenantMapper) ToUpsertTenantFeatureParams(feature *models.TenantFeature) sqlcgen.UpsertTenantFeatureParams {
	params := sqlcgen.UpsertTenantFeatureParams{
		TenantID:  feature.TenantID,
		Key:       feature.Key,
		Value:     feature.Value,
		ValueType: &feature.ValueType,
	}

	if feature.Description != nil {
		params.Description = feature.Description
	}

	return params
}

func (m *TenantMapper) ToProtobufTenantFeature(feature *models.TenantFeature) *tenantpb.TenantFeature {
	return &tenantpb.TenantFeature{
		Key:   feature.Key,
		Value: feature.Value,
	}
}

// TenantConnectionString mapping functions

func (m *TenantMapper) FromSQLCTenantConnectionString(sqlcCS sqlcgen.TenantConnectionString) *models.TenantConnectionString {
	var isEncrypted bool
	if sqlcCS.IsEncrypted != nil {
		isEncrypted = *sqlcCS.IsEncrypted
	}

	return &models.TenantConnectionString{
		ID:          sqlcCS.ID,
		TenantID:    sqlcCS.TenantID,
		Key:         sqlcCS.Key,
		Value:       sqlcCS.Value,
		IsEncrypted: isEncrypted,
		CreatedAt:   sqlcCS.CreatedAt,
		UpdatedAt:   sqlcCS.UpdatedAt,
	}
}

func (m *TenantMapper) ToCreateTenantConnectionStringParams(cs *models.TenantConnectionString) sqlcgen.CreateTenantConnectionStringParams {
	return sqlcgen.CreateTenantConnectionStringParams{
		TenantID:    cs.TenantID,
		Key:         cs.Key,
		Value:       cs.Value,
		IsEncrypted: &cs.IsEncrypted,
	}
}

func (m *TenantMapper) ToProtobufTenantConnectionString(cs *models.TenantConnectionString) *tenantpb.TenantConnectionString {
	return &tenantpb.TenantConnectionString{
		Key:   cs.Key,
		Value: cs.Value,
	}
}

// TenantDomain mapping functions

func (m *TenantMapper) FromSQLCTenantDomain(sqlcDomain sqlcgen.TenantDomain) *models.TenantDomain {
	domain := &models.TenantDomain{
		ID:         sqlcDomain.ID,
		TenantID:   sqlcDomain.TenantID,
		Domain:     sqlcDomain.Domain,
		IsPrimary:  *sqlcDomain.IsPrimary,
		IsVerified: *sqlcDomain.IsVerified,
		CreatedAt:  sqlcDomain.CreatedAt,
		UpdatedAt:  sqlcDomain.UpdatedAt,
	}

	if sqlcDomain.SslCertificate != nil {
		domain.SSLCertificate = sqlcDomain.SslCertificate
	}
	if sqlcDomain.SslPrivateKey != nil {
		domain.SSLPrivateKey = sqlcDomain.SslPrivateKey
	}
	if sqlcDomain.VerificationToken != nil {
		domain.VerificationToken = sqlcDomain.VerificationToken
	}
	if sqlcDomain.VerifiedAt.Valid {
		domain.VerifiedAt = &sqlcDomain.VerifiedAt.Time
	}

	return domain
}

// TenantMetadata mapping functions

func (m *TenantMapper) FromSQLCTenantMetadata(sqlcMetadata sqlcgen.TenantMetadatum) (*models.TenantMetadata, error) {
	metadata := &models.TenantMetadata{
		ID:            sqlcMetadata.ID,
		TenantID:      sqlcMetadata.TenantID,
		SchemaVersion: *sqlcMetadata.SchemaVersion,
		CreatedAt:     sqlcMetadata.CreatedAt,
		UpdatedAt:     sqlcMetadata.UpdatedAt,
	}

	// Parse JSONB data
	if len(sqlcMetadata.Data) > 0 {
		var data map[string]interface{}
		if err := json.Unmarshal(sqlcMetadata.Data, &data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tenant metadata: %w", err)
		}
		metadata.Data = data
	}

	return metadata, nil
}

// Statistics and analytics mapping functions

func (m *TenantMapper) FromSQLCTenantSummary(sqlcSummary sqlcgen.TenantSummary) *models.TenantSummary {
	return &models.TenantSummary{
		ID:           sqlcSummary.ID,
		Name:         sqlcSummary.Name,
		DisplayName:  sqlcSummary.DisplayName,
		Region:       sqlcSummary.Region,
		TenantDB:     sqlcSummary.TenantDb.(int32),
		IsActive:     *sqlcSummary.IsActive,
		CreatedAt:    sqlcSummary.CreatedAt,
		AdminCount:   sqlcSummary.AdminCount,
		DomainCount:  sqlcSummary.DomainCount,
		FeatureCount: sqlcSummary.FeatureCount,
	}
}

func (m *TenantMapper) FromSQLCTenantStats(sqlcStats sqlcgen.GetTenantStatsRow) *models.TenantStats {
	// Create tenant from embedded fields
	tenant := &models.Tenant{
		ID:           sqlcStats.ID,
		Name:         sqlcStats.Name,
		DisplayName:  sqlcStats.DisplayName,
		Region:       sqlcStats.Region,
		TenantDB:     sqlcStats.TenantDb.(int32),
		IsActive:     *sqlcStats.IsActive,
		CreatedAt:    sqlcStats.CreatedAt,
		UpdatedAt:    sqlcStats.UpdatedAt,
		MaxUsers:     *sqlcStats.MaxUsers,
		MaxStorageGB: *sqlcStats.MaxStorageGb,
	}

	if sqlcStats.Logo != nil {
		tenant.Logo = sqlcStats.Logo
	}
	if sqlcStats.CreatedBy.Valid {
		createdBy := uuid.UUID(sqlcStats.CreatedBy.Bytes)
		tenant.CreatedBy = &createdBy
	}
	if sqlcStats.SubscriptionPlan != nil {
		tenant.SubscriptionPlan = sqlcStats.SubscriptionPlan
	}
	if sqlcStats.SubscriptionExpiresAt.Valid {
		tenant.SubscriptionExpiresAt = &sqlcStats.SubscriptionExpiresAt.Time
	}

	return &models.TenantStats{
		Tenant:          tenant,
		VerifiedDomains: sqlcStats.VerifiedDomains,
		FeatureCount:    sqlcStats.FeatureCount,
		RecentActivity:  sqlcStats.RecentActivity,
	}
}

func (m *TenantMapper) FromSQLCTenantsByRegionStats(sqlcStats sqlcgen.GetTenantsByRegionStatsRow) *models.TenantRegionStats {
	return &models.TenantRegionStats{
		Region:              sqlcStats.Region,
		TenantCount:         sqlcStats.TenantCount,
		ActiveCount:         sqlcStats.ActiveCount,
		SeparateDBCount:     sqlcStats.SeparateDbCount,
		SeparateSchemaCount: sqlcStats.SeparateSchemaCount,
	}
}

func (m *TenantMapper) FromSQLCTenantSubscriptionStats(sqlcStats sqlcgen.GetTenantSubscriptionStatsRow) *models.TenantSubscriptionStats {
	var subscriptionPlan string
	if sqlcStats.SubscriptionPlan != nil {
		subscriptionPlan = *sqlcStats.SubscriptionPlan
	}

	return &models.TenantSubscriptionStats{
		SubscriptionPlan: subscriptionPlan,
		TenantCount:      sqlcStats.TenantCount,
		ActiveCount:      sqlcStats.ActiveCount,
		ExpiredCount:     sqlcStats.ExpiredCount,
	}
}
