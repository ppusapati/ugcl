package services

import "go.uber.org/fx"

// Fx modules for different implementations
var ServiceModule = fx.Module("service",
	fx.Provide(
		NewPermissionDefService,
		NewPermissionService,
		NewRoleService,
		fx.Annotate(
			NewUserService,
			fx.As(new(IUserService)),
		),
	),
)
