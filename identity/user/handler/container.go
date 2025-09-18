package handler

import (
	"go.uber.org/fx"
)

type Handlers struct {
	UserHandler          *UserHandler
	RoleHandler          *RoleHandler
	PermissionDefHandler *PermissionDefHandler
	PermissionHandler    *PermissionHandler
}

var HandlerModule = fx.Module("handler",
	fx.Provide(
		NewPermissionDefHandler,
		NewUserHandler,
		NewPermissionHandler,
		NewRoleHandler,
		func(
			uh *UserHandler,
			rh *RoleHandler,
			ph *PermissionHandler,
			pdh *PermissionDefHandler,
		) Handlers {
			return Handlers{
				UserHandler:          uh,
				RoleHandler:          rh,
				PermissionHandler:    ph,
				PermissionDefHandler: pdh,
			}
		},
	))
