package config

import (
	"fmt"
	"time"
)

// ElasticsearchConfig holds Elasticsearch configuration
type ElasticsearchConfig struct {
	// Connection settings
	Addresses []string `mapstructure:"addresses" json:"addresses"`
	Username  string   `mapstructure:"username" json:"username"`
	Password  string   `mapstructure:"password" json:"password"`

	// Connection pool settings
	MaxIdleConns        int           `mapstructure:"max_idle_conns" json:"max_idle_conns"`
	MaxOpenConns        int           `mapstructure:"max_open_conns" json:"max_open_conns"`
	ConnMaxLifetime     time.Duration `mapstructure:"conn_max_lifetime" json:"conn_max_lifetime"`
	ResponseTimeout     time.Duration `mapstructure:"response_timeout" json:"response_timeout"`
	DialTimeout         time.Duration `mapstructure:"dial_timeout" json:"dial_timeout"`
	TLSHandshakeTimeout time.Duration `mapstructure:"tls_handshake_timeout" json:"tls_handshake_timeout"`

	// Retry settings
	MaxRetries     int           `mapstructure:"max_retries" json:"max_retries"`
	RetryBackoff   time.Duration `mapstructure:"retry_backoff" json:"retry_backoff"`
	RetryOnStatus  []int         `mapstructure:"retry_on_status" json:"retry_on_status"`

	// Performance settings
	BulkSize          int           `mapstructure:"bulk_size" json:"bulk_size"`
	BulkTimeout       time.Duration `mapstructure:"bulk_timeout" json:"bulk_timeout"`
	RefreshInterval   string        `mapstructure:"refresh_interval" json:"refresh_interval"`
	MaxResultWindow   int           `mapstructure:"max_result_window" json:"max_result_window"`

	// Index settings
	DefaultShards    int `mapstructure:"default_shards" json:"default_shards"`
	DefaultReplicas  int `mapstructure:"default_replicas" json:"default_replicas"`

	// Security settings
	EnableTLS          bool   `mapstructure:"enable_tls" json:"enable_tls"`
	TLSCertPath        string `mapstructure:"tls_cert_path" json:"tls_cert_path"`
	TLSKeyPath         string `mapstructure:"tls_key_path" json:"tls_key_path"`
	TLSCAPath          string `mapstructure:"tls_ca_path" json:"tls_ca_path"`
	SkipTLSVerify      bool   `mapstructure:"skip_tls_verify" json:"skip_tls_verify"`

	// Monitoring
	EnableMetrics     bool          `mapstructure:"enable_metrics" json:"enable_metrics"`
	MetricsInterval   time.Duration `mapstructure:"metrics_interval" json:"metrics_interval"`
	HealthCheckPath   string        `mapstructure:"health_check_path" json:"health_check_path"`
}

// IndexingConfig holds indexing configuration
type IndexingConfig struct {
	// Worker settings
	WorkerCount       int           `mapstructure:"worker_count" json:"worker_count"`
	QueueSize         int           `mapstructure:"queue_size" json:"queue_size"`
	BatchSize         int           `mapstructure:"batch_size" json:"batch_size"`
	BatchTimeout      time.Duration `mapstructure:"batch_timeout" json:"batch_timeout"`
	ProcessingTimeout time.Duration `mapstructure:"processing_timeout" json:"processing_timeout"`

	// Retry settings
	MaxRetries    int           `mapstructure:"max_retries" json:"max_retries"`
	RetryDelay    time.Duration `mapstructure:"retry_delay" json:"retry_delay"`
	RetryBackoff  float64       `mapstructure:"retry_backoff" json:"retry_backoff"`

	// Index management
	AutoCreateIndices    bool     `mapstructure:"auto_create_indices" json:"auto_create_indices"`
	IndexPrefix          string   `mapstructure:"index_prefix" json:"index_prefix"`
	DefaultMappings      bool     `mapstructure:"default_mappings" json:"default_mappings"`
	EnabledIndexers      []string `mapstructure:"enabled_indexers" json:"enabled_indexers"`

	// Document processing
	MaxDocumentSize      int64  `mapstructure:"max_document_size" json:"max_document_size"`
	EnableTextExtraction bool   `mapstructure:"enable_text_extraction" json:"enable_text_extraction"`
	EnableOCR            bool   `mapstructure:"enable_ocr" json:"enable_ocr"`
	SupportedFormats     []string `mapstructure:"supported_formats" json:"supported_formats"`

	// Performance
	EnableCompression   bool `mapstructure:"enable_compression" json:"enable_compression"`
	EnableAsyncIndexing bool `mapstructure:"enable_async_indexing" json:"enable_async_indexing"`
	CacheEnabled        bool `mapstructure:"cache_enabled" json:"cache_enabled"`
	CacheTTL            time.Duration `mapstructure:"cache_ttl" json:"cache_ttl"`
}

// CacheConfig holds caching configuration
type CacheConfig struct {
	// Redis settings
	RedisAddr     string `mapstructure:"redis_addr" json:"redis_addr"`
	RedisPassword string `mapstructure:"redis_password" json:"redis_password"`
	RedisDB       int    `mapstructure:"redis_db" json:"redis_db"`

	// Connection pool
	PoolSize        int           `mapstructure:"pool_size" json:"pool_size"`
	MinIdleConns    int           `mapstructure:"min_idle_conns" json:"min_idle_conns"`
	MaxConnAge      time.Duration `mapstructure:"max_conn_age" json:"max_conn_age"`
	PoolTimeout     time.Duration `mapstructure:"pool_timeout" json:"pool_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout" json:"idle_timeout"`

	// Cache behavior
	DefaultTTL      time.Duration `mapstructure:"default_ttl" json:"default_ttl"`
	MaxKeySize      int           `mapstructure:"max_key_size" json:"max_key_size"`
	MaxValueSize    int           `mapstructure:"max_value_size" json:"max_value_size"`
	EnableMetrics   bool          `mapstructure:"enable_metrics" json:"enable_metrics"`

	// Prefixes for different cache types
	SearchPrefix    string `mapstructure:"search_prefix" json:"search_prefix"`
	DocumentPrefix  string `mapstructure:"document_prefix" json:"document_prefix"`
	SuggestionPrefix string `mapstructure:"suggestion_prefix" json:"suggestion_prefix"`
	AnalyticsPrefix string `mapstructure:"analytics_prefix" json:"analytics_prefix"`
}

// NewElasticsearchConfig creates a new Elasticsearch configuration with defaults
func NewElasticsearchConfig() *ElasticsearchConfig {
	return &ElasticsearchConfig{
		// Connection defaults
		Addresses: []string{"http://localhost:9200"},
		Username:  "",
		Password:  "",

		// Connection pool defaults
		MaxIdleConns:        10,
		MaxOpenConns:        100,
		ConnMaxLifetime:     30 * time.Minute,
		ResponseTimeout:     30 * time.Second,
		DialTimeout:         30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,

		// Retry defaults
		MaxRetries:    3,
		RetryBackoff:  100 * time.Millisecond,
		RetryOnStatus: []int{502, 503, 504, 429},

		// Performance defaults
		BulkSize:        100,
		BulkTimeout:     5 * time.Second,
		RefreshInterval: "1s",
		MaxResultWindow: 10000,

		// Index defaults
		DefaultShards:   1,
		DefaultReplicas: 1,

		// Security defaults
		EnableTLS:     false,
		SkipTLSVerify: false,

		// Monitoring defaults
		EnableMetrics:   true,
		MetricsInterval: 30 * time.Second,
		HealthCheckPath: "/_cluster/health",
	}
}

// NewIndexingConfig creates a new indexing configuration with defaults
func NewIndexingConfig() *IndexingConfig {
	return &IndexingConfig{
		// Worker defaults
		WorkerCount:       5,
		QueueSize:         10000,
		BatchSize:         100,
		BatchTimeout:      5 * time.Second,
		ProcessingTimeout: 30 * time.Second,

		// Retry defaults
		MaxRetries:   3,
		RetryDelay:   1 * time.Second,
		RetryBackoff: 2.0,

		// Index management defaults
		AutoCreateIndices: true,
		IndexPrefix:       "ugcl",
		DefaultMappings:   true,
		EnabledIndexers:   []string{"document", "form", "user", "notification"},

		// Document processing defaults
		MaxDocumentSize:      50 * 1024 * 1024, // 50MB
		EnableTextExtraction: true,
		EnableOCR:            true,
		SupportedFormats:     []string{"pdf", "doc", "docx", "txt", "rtf", "odt"},

		// Performance defaults
		EnableCompression:   true,
		EnableAsyncIndexing: true,
		CacheEnabled:        true,
		CacheTTL:            1 * time.Hour,
	}
}

// NewCacheConfig creates a new cache configuration with defaults
func NewCacheConfig() *CacheConfig {
	return &CacheConfig{
		// Redis defaults
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       0,

		// Connection pool defaults
		PoolSize:     10,
		MinIdleConns: 5,
		MaxConnAge:   30 * time.Minute,
		PoolTimeout:  5 * time.Second,
		IdleTimeout:  5 * time.Minute,

		// Cache behavior defaults
		DefaultTTL:   1 * time.Hour,
		MaxKeySize:   250,
		MaxValueSize: 1024 * 1024, // 1MB
		EnableMetrics: true,

		// Prefix defaults
		SearchPrefix:     "search:",
		DocumentPrefix:   "doc:",
		SuggestionPrefix: "suggest:",
		AnalyticsPrefix:  "analytics:",
	}
}

// Validate validates the Elasticsearch configuration
func (c *ElasticsearchConfig) Validate() error {
	if len(c.Addresses) == 0 {
		return fmt.Errorf("at least one Elasticsearch address is required")
	}

	if c.MaxRetries < 0 {
		return fmt.Errorf("max_retries cannot be negative")
	}

	if c.BulkSize <= 0 {
		return fmt.Errorf("bulk_size must be positive")
	}

	if c.MaxResultWindow <= 0 {
		return fmt.Errorf("max_result_window must be positive")
	}

	if c.DefaultShards <= 0 {
		return fmt.Errorf("default_shards must be positive")
	}

	if c.DefaultReplicas < 0 {
		return fmt.Errorf("default_replicas cannot be negative")
	}

	return nil
}

// Validate validates the indexing configuration
func (c *IndexingConfig) Validate() error {
	if c.WorkerCount <= 0 {
		return fmt.Errorf("worker_count must be positive")
	}

	if c.QueueSize <= 0 {
		return fmt.Errorf("queue_size must be positive")
	}

	if c.BatchSize <= 0 {
		return fmt.Errorf("batch_size must be positive")
	}

	if c.MaxRetries < 0 {
		return fmt.Errorf("max_retries cannot be negative")
	}

	if c.MaxDocumentSize <= 0 {
		return fmt.Errorf("max_document_size must be positive")
	}

	if c.RetryBackoff < 1.0 {
		return fmt.Errorf("retry_backoff must be >= 1.0")
	}

	return nil
}

// Validate validates the cache configuration
func (c *CacheConfig) Validate() error {
	if c.RedisAddr == "" {
		return fmt.Errorf("redis_addr is required")
	}

	if c.PoolSize <= 0 {
		return fmt.Errorf("pool_size must be positive")
	}

	if c.MaxKeySize <= 0 {
		return fmt.Errorf("max_key_size must be positive")
	}

	if c.MaxValueSize <= 0 {
		return fmt.Errorf("max_value_size must be positive")
	}

	return nil
}

// GetIndexName returns a prefixed index name
func (c *IndexingConfig) GetIndexName(baseName string) string {
	if c.IndexPrefix == "" {
		return baseName
	}
	return fmt.Sprintf("%s_%s", c.IndexPrefix, baseName)
}

// IsIndexerEnabled checks if an indexer is enabled
func (c *IndexingConfig) IsIndexerEnabled(indexerName string) bool {
	for _, enabled := range c.EnabledIndexers {
		if enabled == indexerName {
			return true
		}
	}
	return false
}

// IsSupportedFormat checks if a document format is supported
func (c *IndexingConfig) IsSupportedFormat(format string) bool {
	for _, supported := range c.SupportedFormats {
		if supported == format {
			return true
		}
	}
	return false
}