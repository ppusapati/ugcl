-- name: CreateRelationship :one
INSERT INTO table_relationships (
    name, source_table_id, source_column_id, target_table_id, target_column_id,
    relationship_type, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetRelationship :one
SELECT * FROM table_relationships
WHERE id = $1 AND is_active = true;

-- name: ListRelationshipsByTable :many
SELECT r.*,
       st.table_name as source_table_name, st.display_name as source_table_display_name,
       sc.column_name as source_column_name, sc.display_name as source_column_display_name,
       tt.table_name as target_table_name, tt.display_name as target_table_display_name,
       tc.column_name as target_column_name, tc.display_name as target_column_display_name
FROM table_relationships r
JOIN tables_metadata st ON r.source_table_id = st.id
JOIN columns_metadata sc ON r.source_column_id = sc.id
JOIN tables_metadata tt ON r.target_table_id = tt.id
JOIN columns_metadata tc ON r.target_column_id = tc.id
WHERE (r.source_table_id = $1 OR r.target_table_id = $1) AND r.is_active = true
ORDER BY r.name;

-- name: ListAllRelationships :many
SELECT r.*,
       st.table_name as source_table_name, st.display_name as source_table_display_name,
       sc.column_name as source_column_name, sc.display_name as source_column_display_name,
       tt.table_name as target_table_name, tt.display_name as target_table_display_name,
       tc.column_name as target_column_name, tc.display_name as target_column_display_name,
       ss.schema_name as source_schema_name, ts.schema_name as target_schema_name
FROM table_relationships r
JOIN tables_metadata st ON r.source_table_id = st.id
JOIN schemas_metadata ss ON st.schema_id = ss.id
JOIN columns_metadata sc ON r.source_column_id = sc.id
JOIN tables_metadata tt ON r.target_table_id = tt.id
JOIN schemas_metadata ts ON tt.schema_id = ts.id
JOIN columns_metadata tc ON r.target_column_id = tc.id
WHERE r.is_active = true AND st.is_active = true AND tt.is_active = true
ORDER BY ss.display_name, st.display_name, r.name;

-- name: GetRelationshipsByType :many
SELECT r.*,
       st.table_name as source_table_name, st.display_name as source_table_display_name,
       sc.column_name as source_column_name, sc.display_name as source_column_display_name,
       tt.table_name as target_table_name, tt.display_name as target_table_display_name,
       tc.column_name as target_column_name, tc.display_name as target_column_display_name
FROM table_relationships r
JOIN tables_metadata st ON r.source_table_id = st.id
JOIN columns_metadata sc ON r.source_column_id = sc.id
JOIN tables_metadata tt ON r.target_table_id = tt.id
JOIN columns_metadata tc ON r.target_column_id = tc.id
WHERE r.relationship_type = $1 AND r.is_active = true
ORDER BY r.name;

-- name: UpdateRelationship :one
UPDATE table_relationships
SET name = $2, relationship_type = $3, updated_by = $4, updated_at = NOW()
WHERE id = $1 AND is_active = true
RETURNING *;

-- name: DeleteRelationship :exec
UPDATE table_relationships
SET is_active = false, updated_by = $2, updated_at = NOW()
WHERE id = $1;