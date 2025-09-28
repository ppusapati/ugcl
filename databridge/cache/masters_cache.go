// =============================================================================
// databridge/cache/masters_cache.go - Cache layer for master data
// =============================================================================
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"databridge/services"

	"github.com/google/uuid"
)

// CacheEntry represents a cached item with expiration
type CacheEntry[T any] struct {
	Data      T
	ExpiresAt time.Time
}

// IsExpired checks if the cache entry has expired
func (e *CacheEntry[T]) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// MastersCache provides caching for master data with event-driven invalidation
type MastersCache struct {
	// Separate caches for different data types
	schemas  map[string]*CacheEntry[*services.Schema]
	tables   map[string]*CacheEntry[*services.Table]
	columns  map[string]*CacheEntry[*services.Column]

	// Index caches for faster lookups
	tablesBySchema map[string]*CacheEntry[[]*services.Table]
	columnsByTable map[string]*CacheEntry[[]*services.Column]

	// Mutex for thread-safe access
	mu sync.RWMutex

	// Configuration
	defaultTTL time.Duration

	// Fallback client for cache misses
	mastersClient services.IMastersClient
}

// NewMastersCache creates a new masters cache
func NewMastersCache(mastersClient services.IMastersClient) *MastersCache {
	return &MastersCache{
		schemas:        make(map[string]*CacheEntry[*services.Schema]),
		tables:         make(map[string]*CacheEntry[*services.Table]),
		columns:        make(map[string]*CacheEntry[*services.Column]),
		tablesBySchema: make(map[string]*CacheEntry[[]*services.Table]),
		columnsByTable: make(map[string]*CacheEntry[[]*services.Column]),
		defaultTTL:     30 * time.Minute, // 30 minute TTL
		mastersClient:  mastersClient,
	}
}

// GetSchema retrieves a schema by ID, using cache first, then fallback to masters service
func (c *MastersCache) GetSchema(ctx context.Context, schemaID uuid.UUID) (*services.Schema, error) {
	key := schemaID.String()

	c.mu.RLock()
	if entry, exists := c.schemas[key]; exists && !entry.IsExpired() {
		c.mu.RUnlock()
		return entry.Data, nil
	}
	c.mu.RUnlock()

	// Cache miss or expired - fetch from masters service
	schema, err := c.mastersClient.GetSchemaByID(ctx, schemaID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schema from masters service: %w", err)
	}

	// Cache the result
	c.cacheSchema(key, schema)
	return schema, nil
}

// GetTable retrieves a table by ID, using cache first, then fallback to masters service
func (c *MastersCache) GetTable(ctx context.Context, tableID uuid.UUID) (*services.Table, error) {
	key := tableID.String()

	c.mu.RLock()
	if entry, exists := c.tables[key]; exists && !entry.IsExpired() {
		c.mu.RUnlock()
		return entry.Data, nil
	}
	c.mu.RUnlock()

	// Cache miss or expired - fetch from masters service
	table, err := c.mastersClient.GetTableByID(ctx, tableID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch table from masters service: %w", err)
	}

	// Cache the result
	c.cacheTable(key, table)
	return table, nil
}

// GetTablesBySchema retrieves tables for a schema, using cache first
func (c *MastersCache) GetTablesBySchema(ctx context.Context, schemaID uuid.UUID, importableOnly bool) ([]*services.Table, error) {
	key := fmt.Sprintf("%s:importable:%v", schemaID.String(), importableOnly)

	c.mu.RLock()
	if entry, exists := c.tablesBySchema[key]; exists && !entry.IsExpired() {
		c.mu.RUnlock()
		return entry.Data, nil
	}
	c.mu.RUnlock()

	// Cache miss or expired - fetch from masters service
	tables, err := c.mastersClient.GetTables(ctx, &schemaID, importableOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tables from masters service: %w", err)
	}

	// Cache the result and individual tables
	c.cacheTablesBySchema(key, tables)
	for _, table := range tables {
		c.cacheTable(table.ID, table)
	}

	return tables, nil
}

// GetColumnsByTable retrieves columns for a table, using cache first
func (c *MastersCache) GetColumnsByTable(ctx context.Context, tableID uuid.UUID, importableOnly bool) ([]*services.Column, error) {
	key := fmt.Sprintf("%s:importable:%v", tableID.String(), importableOnly)

	c.mu.RLock()
	if entry, exists := c.columnsByTable[key]; exists && !entry.IsExpired() {
		c.mu.RUnlock()
		return entry.Data, nil
	}
	c.mu.RUnlock()

	// Cache miss or expired - fetch from masters service
	columns, err := c.mastersClient.GetColumns(ctx, tableID, importableOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch columns from masters service: %w", err)
	}

	// Cache the result and individual columns
	c.cacheColumnsByTable(key, columns)
	for _, column := range columns {
		c.cacheColumn(column.ID, column)
	}

	return columns, nil
}

// Event-driven cache invalidation methods

// InvalidateSchema invalidates schema cache and related data
func (c *MastersCache) InvalidateSchema(schemaID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove schema from cache
	delete(c.schemas, schemaID)

	// Remove tables by schema cache
	for key := range c.tablesBySchema {
		if len(key) > 36 && key[:36] == schemaID {
			delete(c.tablesBySchema, key)
		}
	}
}

// InvalidateTable invalidates table cache and related data
func (c *MastersCache) InvalidateTable(tableID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove table from cache
	delete(c.tables, tableID)

	// Remove columns by table cache
	for key := range c.columnsByTable {
		if len(key) > 36 && key[:36] == tableID {
			delete(c.columnsByTable, key)
		}
	}

	// Remove from tables by schema cache (we don't know the schema ID, so clear all)
	c.tablesBySchema = make(map[string]*CacheEntry[[]*services.Table])
}

// InvalidateColumn invalidates column cache and related data
func (c *MastersCache) InvalidateColumn(columnID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove column from cache
	delete(c.columns, columnID)

	// Remove from columns by table cache (we don't know the table ID, so clear all)
	c.columnsByTable = make(map[string]*CacheEntry[[]*services.Column])
}

// UpdateSchema updates schema in cache
func (c *MastersCache) UpdateSchema(schemaData map[string]interface{}) {
	if id, exists := schemaData["id"].(string); exists {
		schema := c.mapToSchema(schemaData)
		if schema != nil {
			c.cacheSchema(id, schema)
		}
	}
}

// UpdateTable updates table in cache
func (c *MastersCache) UpdateTable(tableData map[string]interface{}) {
	if id, exists := tableData["id"].(string); exists {
		table := c.mapToTable(tableData)
		if table != nil {
			c.cacheTable(id, table)
		}
	}
}

// UpdateColumn updates column in cache
func (c *MastersCache) UpdateColumn(columnData map[string]interface{}) {
	if id, exists := columnData["id"].(string); exists {
		column := c.mapToColumn(columnData)
		if column != nil {
			c.cacheColumn(id, column)
		}
	}
}

// Private helper methods

func (c *MastersCache) cacheSchema(key string, schema *services.Schema) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.schemas[key] = &CacheEntry[*services.Schema]{
		Data:      schema,
		ExpiresAt: time.Now().Add(c.defaultTTL),
	}
}

func (c *MastersCache) cacheTable(key string, table *services.Table) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.tables[key] = &CacheEntry[*services.Table]{
		Data:      table,
		ExpiresAt: time.Now().Add(c.defaultTTL),
	}
}

func (c *MastersCache) cacheColumn(key string, column *services.Column) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.columns[key] = &CacheEntry[*services.Column]{
		Data:      column,
		ExpiresAt: time.Now().Add(c.defaultTTL),
	}
}

func (c *MastersCache) cacheTablesBySchema(key string, tables []*services.Table) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.tablesBySchema[key] = &CacheEntry[[]*services.Table]{
		Data:      tables,
		ExpiresAt: time.Now().Add(c.defaultTTL),
	}
}

func (c *MastersCache) cacheColumnsByTable(key string, columns []*services.Column) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.columnsByTable[key] = &CacheEntry[[]*services.Column]{
		Data:      columns,
		ExpiresAt: time.Now().Add(c.defaultTTL),
	}
}

// Conversion helpers
func (c *MastersCache) mapToSchema(data map[string]interface{}) *services.Schema {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil
	}

	var schema services.Schema
	if err := json.Unmarshal(jsonData, &schema); err != nil {
		return nil
	}

	return &schema
}

func (c *MastersCache) mapToTable(data map[string]interface{}) *services.Table {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil
	}

	var table services.Table
	if err := json.Unmarshal(jsonData, &table); err != nil {
		return nil
	}

	return &table
}

func (c *MastersCache) mapToColumn(data map[string]interface{}) *services.Column {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil
	}

	var column services.Column
	if err := json.Unmarshal(jsonData, &column); err != nil {
		return nil
	}

	return &column
}

// Cache statistics and maintenance

// GetStats returns cache statistics
func (c *MastersCache) GetStats() map[string]int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]int{
		"schemas_cached":         len(c.schemas),
		"tables_cached":          len(c.tables),
		"columns_cached":         len(c.columns),
		"tables_by_schema_cached": len(c.tablesBySchema),
		"columns_by_table_cached": len(c.columnsByTable),
	}
}

// CleanExpired removes expired entries from cache
func (c *MastersCache) CleanExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clean schemas
	for key, entry := range c.schemas {
		if entry.IsExpired() {
			delete(c.schemas, key)
		}
	}

	// Clean tables
	for key, entry := range c.tables {
		if entry.IsExpired() {
			delete(c.tables, key)
		}
	}

	// Clean columns
	for key, entry := range c.columns {
		if entry.IsExpired() {
			delete(c.columns, key)
		}
	}

	// Clean table lists
	for key, entry := range c.tablesBySchema {
		if entry.IsExpired() {
			delete(c.tablesBySchema, key)
		}
	}

	// Clean column lists
	for key, entry := range c.columnsByTable {
		if entry.IsExpired() {
			delete(c.columnsByTable, key)
		}
	}
}

// Clear removes all entries from cache
func (c *MastersCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.schemas = make(map[string]*CacheEntry[*services.Schema])
	c.tables = make(map[string]*CacheEntry[*services.Table])
	c.columns = make(map[string]*CacheEntry[*services.Column])
	c.tablesBySchema = make(map[string]*CacheEntry[[]*services.Table])
	c.columnsByTable = make(map[string]*CacheEntry[[]*services.Column])
}