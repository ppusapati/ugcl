package helper

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// BuildJSONtoDBColumnMap builds mapping between JSON field names and DB column names
// This uses your existing implementation which is more robust
func BuildJSONtoDBColumnMap(db *gorm.DB, model any) (map[string]string, error) {
	// Get schema
	stmt := &gorm.Statement{DB: db}
	fmt.Println("model", model)
	if err := stmt.Parse(model); err != nil {
		fmt.Println("error at BuildJSONtoDBColumnMap line 15", err)
		return nil, err
	}
	out := make(map[string]string)
	for _, field := range stmt.Schema.Fields {
		jsonName := field.Tag.Get("json")
		if idx := strings.Index(jsonName, ","); idx != -1 {
			jsonName = jsonName[:idx]
		}
		dbCol := field.DBName
		if jsonName != "" && dbCol != "" && jsonName != "-" {
			out[jsonName] = dbCol
		}
	}

	return out, nil
}
