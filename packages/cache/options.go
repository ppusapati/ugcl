package cache

import (
	"time"

	"p9e.in/ugcl/packages/api/v1/config"
)

// CacheOptions defines configuration for the cache
type CacheOptions struct {
	Enabled             bool
	EnableExpiry        bool
	DefaultTTL          time.Duration
	ExpiryCheckInterval time.Duration
	MaxEntries          *int
}

// Option represents a cache option
type Option func(*CacheOptions)

// defaultOptions provides sensible default cache configuration
func defaultOptions() CacheOptions {
	return CacheOptions{
		Enabled:             true,
		EnableExpiry:        true,
		DefaultTTL:          1 * time.Hour,
		ExpiryCheckInterval: 5 * time.Minute,
		MaxEntries:          new(int),
	}
}

// configureOptions sets the cache configuration from a config object or uses defaults
func configureOptions(cfg *config.Data_Cache) CacheOptions {
	options := defaultOptions()

	if cfg != nil {
		if cfg.Enabled != nil {
			options.Enabled = cfg.Enabled.GetValue()
		}
		if cfg.MaxEntries != nil {
			*options.MaxEntries = int(cfg.MaxEntries.GetValue())
		}
		if cfg.DefaultTtl != nil {
			options.DefaultTTL = cfg.DefaultTtl.AsDuration()
		}
		if cfg.EnableExpiry != nil {
			options.EnableExpiry = cfg.EnableExpiry.GetValue()
		}
		if cfg.ExpiryCheckInterval != nil {
			options.ExpiryCheckInterval = cfg.ExpiryCheckInterval.AsDuration()
		}
	}

	return options
}

// WithEnabled sets whether the cache is enabled
func WithEnabled(enabled bool) Option {
	return func(opts *CacheOptions) {
		opts.Enabled = enabled
	}
}

// WithExpiry enables cache expiry and sets TTL
func WithExpiry(ttl time.Duration) Option {
	return func(opts *CacheOptions) {
		opts.EnableExpiry = true
		opts.DefaultTTL = ttl
	}
}

// WithExpiryCheckInterval sets how often expired entries are checked
func WithExpiryCheckInterval(interval time.Duration) Option {
	return func(opts *CacheOptions) {
		opts.ExpiryCheckInterval = interval
	}
}

// WithMaxEntries sets the maximum number of entries in the cache
func WithMaxEntries(maxEntries int) Option {
	return func(opts *CacheOptions) {
		opts.MaxEntries = &maxEntries
	}
}
