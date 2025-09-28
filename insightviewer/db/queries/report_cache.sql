-- =============================================================================
-- Report Cache Queries - InsightViewer Service
-- =============================================================================

-- name: CreateCacheEntry :one
INSERT INTO report_cache (
    cache_key,
    cache_type,
    report_id,
    data,
    data_text,
    size_bytes,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING
    id,
    cache_key,
    cache_type,
    report_id,
    data,
    data_text,
    size_bytes,
    created_at,
    expires_at,
    last_accessed_at,
    access_count,
    hit_count,
    miss_count;

-- name: GetCacheEntryByKey :one
SELECT
    id,
    cache_key,
    cache_type,
    report_id,
    data,
    data_text,
    size_bytes,
    created_at,
    expires_at,
    last_accessed_at,
    access_count,
    hit_count,
    miss_count
FROM report_cache
WHERE cache_key = $1 AND (expires_at IS NULL OR expires_at > NOW());

-- name: UpdateCacheAccess :exec
UPDATE report_cache
SET
    last_accessed_at = NOW(),
    access_count = access_count + 1,
    hit_count = hit_count + 1
WHERE cache_key = $1;

-- name: IncrementCacheMiss :exec
UPDATE report_cache
SET
    miss_count = miss_count + 1
WHERE cache_key = $1;

-- name: DeleteCacheEntry :exec
DELETE FROM report_cache
WHERE cache_key = $1;

-- name: DeleteExpiredCache :exec
DELETE FROM report_cache
WHERE expires_at IS NOT NULL AND expires_at <= NOW();

-- name: DeleteCacheByReportID :exec
DELETE FROM report_cache
WHERE report_id = $1;

-- name: GetCacheStatistics :one
SELECT
    COUNT(*) as total_entries,
    SUM(size_bytes) as total_size_bytes,
    AVG(size_bytes) as avg_size_bytes,
    SUM(hit_count) as total_hits,
    SUM(miss_count) as total_misses,
    CASE
        WHEN SUM(hit_count + miss_count) > 0
        THEN (SUM(hit_count)::float / SUM(hit_count + miss_count)::float) * 100
        ELSE 0
    END as hit_rate_percent,
    COUNT(CASE WHEN cache_type = 'result' THEN 1 END) as result_entries,
    COUNT(CASE WHEN cache_type = 'metadata' THEN 1 END) as metadata_entries,
    COUNT(CASE WHEN cache_type = 'query_plan' THEN 1 END) as query_plan_entries,
    COUNT(CASE WHEN cache_type = 'field_values' THEN 1 END) as field_values_entries
FROM report_cache
WHERE created_at >= $1;

-- name: GetLeastRecentlyUsedCache :many
SELECT
    id,
    cache_key,
    cache_type,
    report_id,
    size_bytes,
    created_at,
    last_accessed_at,
    access_count
FROM report_cache
ORDER BY last_accessed_at ASC
LIMIT $1;

-- name: GetLargestCacheEntries :many
SELECT
    id,
    cache_key,
    cache_type,
    report_id,
    size_bytes,
    created_at,
    last_accessed_at,
    access_count
FROM report_cache
ORDER BY size_bytes DESC
LIMIT $1;

-- name: CleanupCacheBySize :exec
DELETE FROM report_cache
WHERE id IN (
    SELECT id
    FROM report_cache
    ORDER BY last_accessed_at ASC
    LIMIT $1
);

-- name: GetCacheEntriesByType :many
SELECT
    id,
    cache_key,
    cache_type,
    report_id,
    data,
    data_text,
    size_bytes,
    created_at,
    expires_at,
    last_accessed_at,
    access_count,
    hit_count,
    miss_count
FROM report_cache
WHERE cache_type = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;