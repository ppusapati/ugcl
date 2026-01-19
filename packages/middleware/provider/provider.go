package provider

import (
	"p9e.in/ugcl/packages/middleware/dbmiddleware"

	"github.com/google/wire"
)

var MiddlewareSet = wire.NewSet(
	dbmiddleware.NewDBResolver,
)
