package config

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
	"p9e.in/ugcl/models"
)

func Migrations(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "26062025_create_tables",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.User{}, &models.DairySite{}, &models.DprSite{}, &models.Contractor{},
					&models.Mnr{}, &models.Material{}, &models.Payment{}, &models.Diesel{}, &models.Eway{}, &models.Painting{},
					&models.Stock{}, &models.Water{}, &models.Wrapping{}, &models.Task{}, &models.Nmr_Vehicle{}, &models.VehicleLog{})
			},
			// Rollback: func(tx *gorm.DB) error {
			// 	return tx.Migrator().DropTable("dairy_sites")
			// },
		},
	})

	return m.Migrate()
}
