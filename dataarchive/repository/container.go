package repository

import (
	"database/sql"

	"p9e.in/ugcl/dataarchive/repository/interfaces"
	"p9e.in/ugcl/dataarchive/repository/sqlc"
)

type Container struct {
	dataArchiveRepo interfaces.DataArchiveRepository
}

func NewContainer(db *sql.DB) *Container {
	return &Container{
		dataArchiveRepo: sqlc.NewDataArchiveRepository(db),
	}
}

func (c *Container) GetDataArchiveRepository() interfaces.DataArchiveRepository {
	return c.dataArchiveRepo
}