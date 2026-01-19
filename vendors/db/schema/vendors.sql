-- schema.sql

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS vendors (
    id         BIGSERIAL PRIMARY KEY,
    uuid       UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    company_name VARCHAR(255) NOT NULL,
    company_type VARCHAR(100),
    gst VARCHAR(50),
    pan VARCHAR(50) UNIQUE,
    vendor_category VARCHAR(100), -- RAW_MATERIAL, EQUIPMENT, SERVICES, CONSUMABLES
    person_id BIGINT REFERENCES users(id),
    payment_terms TEXT,
    credit_period_days INTEGER DEFAULT 0,
    bank_name VARCHAR(255),
    account_number VARCHAR(100),
    ifsc VARCHAR(20),
    rating INTEGER DEFAULT 0 CHECK (rating >= 0 AND rating <= 5),
    is_blacklisted BOOLEAN DEFAULT FALSE,
    contracts TEXT[], -- array of contract IDs
    purchase_orders TEXT[], -- array of PO IDs
    status VARCHAR(50) DEFAULT 'ACTIVE', -- ACTIVE, INACTIVE, SUSPENDED, BLACKLISTED
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT chk_rating CHECK (rating >= 0 AND rating <= 5),
    CONSTRAINT chk_credit_period CHECK (credit_period_days >= 0)
);

-- Indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_vendors_person_id ON vendors(person_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vendors_status ON vendors(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vendors_company ON vendors(company_name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vendors_category ON vendors(vendor_category) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vendors_pan ON vendors(pan) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vendors_gst ON vendors(gst) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vendors_rating ON vendors(rating) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vendors_blacklisted ON vendors(is_blacklisted) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vendors_contracts ON vendors USING GIN(contracts) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vendors_purchase_orders ON vendors USING GIN(purchase_orders) WHERE deleted_at IS NULL;

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'update_vendors_updated_at'
    ) THEN
        CREATE TRIGGER update_vendors_updated_at
            BEFORE UPDATE ON vendors
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;
