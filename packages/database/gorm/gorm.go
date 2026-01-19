package gorm

import (
	"fmt"
	"time"

	"p9e.in/ugcl/packages/api/v1/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	maxOpenConns    = 60
	connMaxLifetime = 120
	maxIdleConns    = 30
	connMaxIdleTime = 20
)

// Return new Postgresql db instance
func NewPsqlDB(c *config.Data) (*gorm.DB, func(), error) {
	dataSourceName := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable password=%s",
		c.Postgres.Host,
		c.Postgres.Port,
		c.Postgres.User,
		c.Postgres.Dbname,
		c.Postgres.Password,
	)
	db, err := gorm.Open(postgres.Open(dataSourceName), &gorm.Config{SkipDefaultTransaction: true, PrepareStmt: true})
	//sqlx.Connect(c.Postgres.PgDriver, dataSourceName)
	if err != nil {
		return nil, nil, err
	}
	sqldb, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	sqldb.SetMaxOpenConns(maxOpenConns)
	sqldb.SetConnMaxLifetime(connMaxLifetime * time.Second)
	sqldb.SetMaxIdleConns(maxIdleConns)
	sqldb.SetConnMaxIdleTime(connMaxIdleTime * time.Second)
	if err = sqldb.Ping(); err != nil {
		return nil, nil, err
	}
	// p9log.Info("Total Number of Database Connections:", sqldb.Stats().OpenConnections)
	Cleanup := func() {
		sqldb.Close()
	}
	return db, Cleanup, nil
}
