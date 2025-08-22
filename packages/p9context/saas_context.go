package p9context

import (
	"context"
	"fmt"

	"p9e.in/ugcl/packages/saas"
)

type (
	currentTenantCtx struct{}
	tenantResolveRes struct{}
	// tenantConfigKey  string
)

func NewCurrentTenant(ctx context.Context, id, name string) context.Context {
	return NewCurrentTenantInfo(ctx, saas.NewBasicTenantInfo(id, name))
}

func NewCurrentTenantInfo(ctx context.Context, info saas.TenantInfo) context.Context {
	fmt.Println("hey buddy: ", info)
	return context.WithValue(ctx, currentTenantCtx{}, info)
}

func FromCurrentTenant(ctx context.Context) (saas.TenantInfo, bool) {
	fmt.Println(ctx)
	value, ok := ctx.Value(currentTenantCtx{}).(saas.TenantInfo)
	if ok {
		return value, true
	}
	return saas.NewBasicTenantInfo("", ""), false
}

func NewTenantResolveRes(ctx context.Context, t *saas.TenantResolveResult) context.Context {
	return context.WithValue(ctx, tenantResolveRes{}, t)
}

func FromTenantResolveRes(ctx context.Context) *saas.TenantResolveResult {
	v, ok := ctx.Value(tenantResolveRes{}).(*saas.TenantResolveResult)
	if ok {
		return v
	}
	return nil
}

// func NewTenantConfigContext(ctx context.Context, tenantId string, cfg *TenantConfig) context.Context {
// 	return context.WithValue(ctx, tenantConfigKey(tenantId), cfg)
// }

// func FromTenantConfigContext(ctx context.Context, tenantId string) (*TenantConfig, bool) {
// 	v, ok := ctx.Value(tenantConfigKey(tenantId)).(*TenantConfig)
// 	if ok {
// 		return v, ok && v != nil
// 	}
// 	return nil, false
// }
