package config

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"strings"

// 	"github.com/go-gormigrate/gormigrate/v2"
// 	"go.uber.org/fx"
// 	"gorm.io/gorm"

// 	pmodels "p9e.in/ugcl/masters/pipeline/models"
// )

// // --- helpers ---------------------------------------------------------------

// func ensureExtensionPgcrypto(tx *gorm.DB) error {
// 	return tx.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto`).Error
// }

// func ensureEnum(tx *gorm.DB, typeName string, values []string) error {
// 	// e.g., "'HDPE','DI','MS'"
// 	valuesList := "'" + strings.Join(values, `','`) + "'"

// 	// Create the enum type if it doesn't exist (note the $$ ... $$ around inner EXECUTE)
// 	create := fmt.Sprintf(`DO $do$
// BEGIN
//     IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = '%s') THEN
//         EXECUTE $$CREATE TYPE %s AS ENUM (%s)$$;
//     END IF;
// END
// $do$;`, typeName, typeName, valuesList)

// 	if err := tx.Exec(create).Error; err != nil {
// 		return err
// 	}

// 	// Add any missing labels idempotently (also dollar-quoted)
// 	for _, v := range values {
// 		add := fmt.Sprintf(`DO $do$
// BEGIN
//     IF NOT EXISTS (
//         SELECT 1
//         FROM pg_enum e
//         JOIN pg_type t ON t.oid = e.enumtypid
//         WHERE t.typname = '%s' AND e.enumlabel = '%s'
//     ) THEN
//         EXECUTE $$ALTER TYPE %s ADD VALUE '%s'$$;
//     END IF;
// END
// $do$;`, typeName, v, typeName, v)

// 		if err := tx.Exec(add).Error; err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// // --- migrations ------------------------------------------------------------

// func RunMigrations(lc fx.Lifecycle, db *gorm.DB) {
// 	lc.Append(fx.Hook{
// 		OnStart: func(ctx context.Context) error {
// 			// Update ID to your preferred convention; must be unique & increasing
// 			migrations := []*gormigrate.Migration{
// 				{
// 					ID: "20250809_create_pipeline_tables",
// 					Migrate: func(tx *gorm.DB) error {
// 						// 1) Extensions / enums first
// 						if err := ensureExtensionPgcrypto(tx); err != nil {
// 							return err
// 						}

// 						// TODO: change these to your real values
// 						pipeTypeEnumValues := []string{"HDPE", "DI", "MS"}
// 						if err := ensureEnum(tx, "pipe_type_enum", pipeTypeEnumValues); err != nil {
// 							return err
// 						}

// 						// 2) Now it's safe to create tables that reference the enum
// 						return tx.AutoMigrate(
// 							&pmodels.Site{},
// 							&pmodels.Zone{},
// 							&pmodels.Pipe{},
// 							&pmodels.Node{},
// 							&pmodels.PipelineSegment{},
// 						)
// 					},
// 					Rollback: func(tx *gorm.DB) error {
// 						// Drop in reverse dependency order
// 						if err := tx.Exec(`DROP TABLE IF EXISTS "pipeline_segments" CASCADE`).Error; err != nil {
// 							return err
// 						}
// 						if err := tx.Exec(`DROP TABLE IF EXISTS "nodes" CASCADE`).Error; err != nil {
// 							return err
// 						}
// 						if err := tx.Exec(`DROP TABLE IF EXISTS "pipes" CASCADE`).Error; err != nil {
// 							return err
// 						}
// 						if err := tx.Exec(`DROP TABLE IF EXISTS "zones" CASCADE`).Error; err != nil {
// 							return err
// 						}
// 						if err := tx.Exec(`DROP TABLE IF EXISTS "sites" CASCADE`).Error; err != nil {
// 							return err
// 						}
// 						// Only drop the enum if you’re sure no other tables use it
// 						if err := tx.Exec(`DO $$
// BEGIN
// 	IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'pipe_type_enum') THEN
// 		EXECUTE 'DROP TYPE pipe_type_enum';
// 	END IF;
// END$$;`).Error; err != nil {
// 							return err
// 						}
// 						return nil
// 					},
// 				},
// 			}

// 			m := gormigrate.New(db, gormigrate.DefaultOptions, migrations)

// 			log.Println("Running DB migrations...")
// 			if err := m.Migrate(); err != nil {
// 				log.Fatalf("could not run migrations: %v", err)
// 			}
// 			log.Println("DB migrations complete.")
// 			return nil
// 		},
// 	})
// }
