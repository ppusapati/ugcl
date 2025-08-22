package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	formbuilderv2 "p9e.in/ugcl/formbuilder/api/v2/form_builder"
)

// FieldOptionsCache manages caching of dropdown options
type FieldOptionsCache struct {
	db *sql.DB
}

func NewFieldOptionsCache(db *sql.DB) *FieldOptionsCache {
	return &FieldOptionsCache{db: db}
}

func (foc *FieldOptionsCache) Get(fieldId string, context map[string]string) ([]*formbuilderv2.FieldOption, error) {
	cacheKey := foc.buildCacheKey(fieldId, context)

	var optionsJSON []byte
	query := `
		SELECT options_data 
		FROM form_options_cache 
		WHERE cache_key = $1 AND expires_at > NOW()
	`

	err := foc.db.QueryRow(query, cacheKey).Scan(&optionsJSON)
	if err != nil {
		return nil, err
	}

	var options []*formbuilderv2.FieldOption
	err = json.Unmarshal(optionsJSON, &options)
	return options, err
}

func (foc *FieldOptionsCache) Set(
	fieldId string,
	context map[string]string,
	options []*formbuilderv2.FieldOption,
	ttlSeconds int32,
) error {
	cacheKey := foc.buildCacheKey(fieldId, context)
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO form_options_cache (source_endpoint, cache_key, options_data, expires_at)
		VALUES ($1, $2, $3, NOW() + INTERVAL '%d seconds')
		ON CONFLICT (source_endpoint, cache_key) 
		DO UPDATE SET options_data = EXCLUDED.options_data, expires_at = EXCLUDED.expires_at
	`

	_, err = foc.db.Exec(fmt.Sprintf(query, ttlSeconds), fieldId, cacheKey, optionsJSON)
	return err
}

func (foc *FieldOptionsCache) buildCacheKey(fieldId string, context map[string]string) string {
	// Build deterministic cache key
	keys := make([]string, 0, len(context))
	for k := range context {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := []string{fieldId}
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s:%s", k, context[k]))
	}

	return strings.Join(parts, "|")
}
